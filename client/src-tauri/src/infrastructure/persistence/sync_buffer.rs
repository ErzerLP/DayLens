//! # SQLite 离线同步缓冲
//!
//! 当网络不可用时，将活动数据缓冲到本地 SQLite。
//! 网络恢复后自动重试上报。

use rusqlite::{params, Connection};
use std::path::PathBuf;
use std::sync::Mutex;

use crate::application::ports::SyncBuffer;
use crate::domain::sync::entity::SyncTask;
use crate::shared::error::{AppError, Result};

// ===== SQL =====

const CREATE_TABLE: &str = "
CREATE TABLE IF NOT EXISTS sync_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    payload TEXT NOT NULL,
    screenshot_path TEXT,
    created_at INTEGER NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    client_ts INTEGER NOT NULL DEFAULT 0
)";

// ===== SqliteSyncBuffer =====

/// SQLite 离线缓冲 — `SyncBuffer` 端口的 SQLite 实现
///
/// 使用本地 SQLite 数据库缓冲失败的上报请求。
pub struct SqliteSyncBuffer {
    /// SQLite 连接（Mutex 保证线程安全）
    conn: Mutex<Connection>,
}

impl SqliteSyncBuffer {
    /// 创建 SQLite 缓冲
    pub fn new(db_path: PathBuf) -> Result<Self> {
        if let Some(parent) = db_path.parent() {
            std::fs::create_dir_all(parent)?;
        }

        let conn = Connection::open(&db_path)
            .map_err(|e| AppError::Database(format!(
                "打开同步数据库失败: {e}",
            )))?;

        conn.execute(CREATE_TABLE, [])
            .map_err(|e| AppError::Database(format!(
                "创建同步表失败: {e}",
            )))?;

        Ok(Self {
            conn: Mutex::new(conn),
        })
    }

    /// 创建内存数据库（用于测试）
    #[cfg(test)]
    pub fn in_memory() -> Result<Self> {
        let conn = Connection::open_in_memory()
            .map_err(|e| AppError::Database(format!(
                "创建内存数据库失败: {e}",
            )))?;

        conn.execute(CREATE_TABLE, [])
            .map_err(|e| AppError::Database(format!(
                "创建同步表失败: {e}",
            )))?;

        Ok(Self {
            conn: Mutex::new(conn),
        })
    }
}

impl SyncBuffer for SqliteSyncBuffer {
    fn enqueue(&self, task: &SyncTask) -> Result<()> {
        let conn = self.conn.lock().map_err(|e| {
            AppError::Database(format!("锁获取失败: {e}"))
        })?;

        conn.execute(
            "INSERT INTO sync_tasks
             (payload, screenshot_path, created_at, client_ts)
             VALUES (?1, ?2, ?3, ?4)",
            params![
                task.payload,
                task.screenshot_path,
                task.created_at,
                task.client_ts,
            ],
        )
        .map_err(|e| AppError::Database(format!(
            "入队失败: {e}",
        )))?;

        Ok(())
    }

    fn pending_tasks(&self, limit: i32) -> Result<Vec<SyncTask>> {
        let conn = self.conn.lock().map_err(|e| {
            AppError::Database(format!("锁获取失败: {e}"))
        })?;

        let mut stmt = conn
            .prepare(
                "SELECT id, payload, screenshot_path, created_at,
                        retry_count, last_error, client_ts
                 FROM sync_tasks
                 ORDER BY created_at ASC
                 LIMIT ?1",
            )
            .map_err(|e| AppError::Database(format!(
                "准备查询失败: {e}",
            )))?;

        let tasks = stmt
            .query_map(params![limit], |row| {
                Ok(SyncTask {
                    id: row.get(0)?,
                    payload: row.get(1)?,
                    screenshot_path: row.get(2)?,
                    created_at: row.get(3)?,
                    retry_count: row.get(4)?,
                    last_error: row.get(5)?,
                    client_ts: row.get(6)?,
                })
            })
            .map_err(|e| AppError::Database(format!(
                "查询失败: {e}",
            )))?
            .filter_map(|r| r.ok())
            .collect();

        Ok(tasks)
    }

    fn mark_completed(&self, task_id: &str) -> Result<()> {
        let conn = self.conn.lock().map_err(|e| {
            AppError::Database(format!("锁获取失败: {e}"))
        })?;

        let id: i64 = task_id.parse().unwrap_or(0);
        conn.execute(
            "DELETE FROM sync_tasks WHERE id = ?1",
            params![id],
        )
        .map_err(|e| AppError::Database(format!(
            "删除失败: {e}",
        )))?;

        Ok(())
    }

    fn increment_retry(&self, task_id: &str) -> Result<()> {
        let conn = self.conn.lock().map_err(|e| {
            AppError::Database(format!("锁获取失败: {e}"))
        })?;

        let id: i64 = task_id.parse().unwrap_or(0);
        conn.execute(
            "UPDATE sync_tasks
             SET retry_count = retry_count + 1
             WHERE id = ?1",
            params![id],
        )
        .map_err(|e| AppError::Database(format!(
            "更新重试计数失败: {e}",
        )))?;

        Ok(())
    }

    fn queue_size(&self) -> Result<i64> {
        let conn = self.conn.lock().map_err(|e| {
            AppError::Database(format!("锁获取失败: {e}"))
        })?;

        let count: i64 = conn
            .query_row(
                "SELECT COUNT(*) FROM sync_tasks",
                [],
                |row| row.get(0),
            )
            .map_err(|e| AppError::Database(format!(
                "统计失败: {e}",
            )))?;

        Ok(count)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 入队和查询() {
        let buffer = SqliteSyncBuffer::in_memory().unwrap();
        let task = SyncTask::new(
            r#"{"app_name":"Code"}"#.to_string(),
            None,
            1234567890,
        );

        buffer.enqueue(&task).unwrap();
        let tasks = buffer.pending_tasks(10).unwrap();

        assert_eq!(tasks.len(), 1);
        assert_eq!(tasks[0].payload, r#"{"app_name":"Code"}"#);
    }

    #[test]
    fn 标记完成后应删除() {
        let buffer = SqliteSyncBuffer::in_memory().unwrap();
        let task = SyncTask::new(
            "payload".to_string(),
            None,
            1234567890,
        );

        buffer.enqueue(&task).unwrap();
        assert_eq!(buffer.queue_size().unwrap(), 1);

        let tasks = buffer.pending_tasks(1).unwrap();
        let id = tasks[0].id.to_string();
        buffer.mark_completed(&id).unwrap();
        assert_eq!(buffer.queue_size().unwrap(), 0);
    }

    #[test]
    fn 重试计数应递增() {
        let buffer = SqliteSyncBuffer::in_memory().unwrap();
        let task = SyncTask::new(
            "payload".to_string(),
            None,
            1234567890,
        );

        buffer.enqueue(&task).unwrap();
        let tasks = buffer.pending_tasks(1).unwrap();
        let id = tasks[0].id.to_string();

        buffer.increment_retry(&id).unwrap();
        buffer.increment_retry(&id).unwrap();

        let tasks = buffer.pending_tasks(10).unwrap();
        assert_eq!(tasks[0].retry_count, 2);
    }

    #[test]
    fn 队列大小统计() {
        let buffer = SqliteSyncBuffer::in_memory().unwrap();

        assert_eq!(buffer.queue_size().unwrap(), 0);

        for i in 0..5 {
            let task = SyncTask::new(
                format!("payload_{i}"),
                None,
                1234567890 + i,
            );
            buffer.enqueue(&task).unwrap();
        }

        assert_eq!(buffer.queue_size().unwrap(), 5);
    }

    #[test]
    fn 待重试任务应按时间排序() {
        let buffer = SqliteSyncBuffer::in_memory().unwrap();

        for i in (0..3).rev() {
            let task = SyncTask::new(
                format!("task_{i}"),
                None,
                1000 + i,
            );
            buffer.enqueue(&task).unwrap();
        }

        let tasks = buffer.pending_tasks(10).unwrap();
        assert_eq!(tasks.len(), 3);
        assert!(tasks[0].created_at <= tasks[1].created_at);
        assert!(tasks[1].created_at <= tasks[2].created_at);
    }
}

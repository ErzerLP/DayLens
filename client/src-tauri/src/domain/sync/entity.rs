//! # 同步领域实体
//!
//! 定义离线缓冲队列中的同步任务数据结构。

use serde::{Deserialize, Serialize};

/// 同步任务 — 待上报的活动数据（存储在 SQLite 缓冲队列中）
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SyncTask {
    /// 本地 ID（SQLite 自增）
    pub id: i64,
    /// 活动数据 JSON
    pub payload: String,
    /// 关联的本地截图路径（待上传）
    pub screenshot_path: Option<String>,
    /// 入队时间戳
    pub created_at: i64,
    /// 重试次数
    pub retry_count: u32,
    /// 最后一次错误信息
    pub last_error: Option<String>,
    /// 客户端采集时间戳（幂等去重键）
    pub client_ts: i64,
}

impl SyncTask {
    /// 创建新的同步任务
    pub fn new(
        payload: String,
        screenshot_path: Option<String>,
        client_ts: i64,
    ) -> Self {
        Self {
            id: 0,
            payload,
            screenshot_path,
            created_at: chrono::Local::now().timestamp(),
            retry_count: 0,
            last_error: None,
            client_ts,
        }
    }
}

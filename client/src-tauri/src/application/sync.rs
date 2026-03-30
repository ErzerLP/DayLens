//! # 同步协调器
//!
//! 周期性重试离线缓冲中的失败任务。

use std::sync::Arc;

use crate::application::ports::{ActivityReporter, SyncBuffer};
use crate::domain::sync::policy::RetryPolicy;

/// 同步协调器
pub struct SyncCoordinator {
    reporter: Arc<dyn ActivityReporter>,
    buffer: Arc<dyn SyncBuffer>,
}

impl SyncCoordinator {
    pub fn new(
        reporter: Arc<dyn ActivityReporter>,
        buffer: Arc<dyn SyncBuffer>,
    ) -> Self {
        Self { reporter, buffer }
    }

    /// 执行一次同步尝试：从缓冲取出任务并重试上报
    pub async fn sync_once(&self) -> u32 {
        let tasks = match self.buffer.pending_tasks(10) {
            Ok(t) => t,
            Err(e) => {
                log::warn!("读取缓冲队列失败: {e}");
                return 0;
            }
        };

        if tasks.is_empty() {
            return 0;
        }

        // 先检查服务端是否可用
        if !self.reporter.is_server_available().await {
            log::debug!("服务端不可用，跳过同步");
            return 0;
        }

        let mut synced = 0u32;

        for task in &tasks {
            // 超过最大重试次数则丢弃
            if RetryPolicy::is_exhausted(task.retry_count) {
                log::warn!(
                    "任务 {} 超过最大重试次数，丢弃",
                    task.id,
                );
                let _ = self.buffer.mark_completed(
                    &task.id.to_string(),
                );
                continue;
            }

            // 尝试解析 payload 并重新上报
            match serde_json::from_str(&task.payload) {
                Ok(activity) => {
                    match self.reporter.report(&activity, None).await {
                        Ok(_) => {
                            let _ = self.buffer.mark_completed(
                                &task.id.to_string(),
                            );
                            synced += 1;
                        }
                        Err(e) => {
                            log::warn!("重试上报失败: {e}");
                            let _ = self.buffer.increment_retry(
                                &task.id.to_string(),
                            );
                        }
                    }
                }
                Err(e) => {
                    log::error!("缓冲 payload 解析失败: {e}");
                    let _ = self.buffer.mark_completed(
                        &task.id.to_string(),
                    );
                }
            }
        }

        if synced > 0 {
            log::info!("同步完成: {synced}/{} 任务成功", tasks.len());
        }

        synced
    }

    pub async fn queue_size(&self) -> i64 {
        self.buffer.queue_size().unwrap_or(0)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::Mutex;
    use crate::domain::activity::entity::Activity;
    use crate::domain::sync::entity::SyncTask;

    // ===== Mock 实现 =====

    struct MockReporter {
        available: bool,
        report_result: Mutex<Vec<crate::shared::error::Result<i64>>>,
    }

    impl MockReporter {
        fn always_ok() -> Self {
            Self {
                available: true,
                report_result: Mutex::new(Vec::new()),
            }
        }
        fn unavailable() -> Self {
            Self {
                available: false,
                report_result: Mutex::new(Vec::new()),
            }
        }
        fn with_results(results: Vec<crate::shared::error::Result<i64>>) -> Self {
            Self {
                available: true,
                report_result: Mutex::new(results),
            }
        }
    }

    #[async_trait::async_trait]
    impl ActivityReporter for MockReporter {
        async fn report(&self, _: &Activity, _: Option<&std::path::Path>) -> crate::shared::error::Result<i64> {
            let mut results = self.report_result.lock().unwrap();
            if results.is_empty() {
                Ok(1)
            } else {
                results.remove(0)
            }
        }
        async fn batch_report(&self, _: &[Activity]) -> crate::shared::error::Result<Vec<i64>> {
            Ok(vec![])
        }
        async fn is_server_available(&self) -> bool {
            self.available
        }
    }

    struct MockBuffer {
        tasks: Mutex<Vec<SyncTask>>,
        completed: Mutex<Vec<String>>,
        retried: Mutex<Vec<String>>,
    }

    impl MockBuffer {
        fn with_tasks(tasks: Vec<SyncTask>) -> Self {
            Self {
                tasks: Mutex::new(tasks),
                completed: Mutex::new(Vec::new()),
                retried: Mutex::new(Vec::new()),
            }
        }
        fn empty() -> Self {
            Self::with_tasks(Vec::new())
        }
    }

    impl SyncBuffer for MockBuffer {
        fn enqueue(&self, task: &SyncTask) -> crate::shared::error::Result<()> {
            self.tasks.lock().unwrap().push(task.clone());
            Ok(())
        }
        fn pending_tasks(&self, _limit: i32) -> crate::shared::error::Result<Vec<SyncTask>> {
            Ok(self.tasks.lock().unwrap().clone())
        }
        fn mark_completed(&self, task_id: &str) -> crate::shared::error::Result<()> {
            self.completed.lock().unwrap().push(task_id.to_string());
            Ok(())
        }
        fn increment_retry(&self, task_id: &str) -> crate::shared::error::Result<()> {
            self.retried.lock().unwrap().push(task_id.to_string());
            Ok(())
        }
        fn queue_size(&self) -> crate::shared::error::Result<i64> {
            Ok(self.tasks.lock().unwrap().len() as i64)
        }
    }

    fn make_valid_task(id: i64, retry_count: u32) -> SyncTask {
        let payload = serde_json::json!({
            "id": 0,
            "clientId": "test",
            "clientTs": 1000,
            "timestamp": 1000,
            "appName": "VSCode",
            "windowTitle": "test.rs",
            "category": "coding",
            "duration": 30,
        });
        SyncTask {
            id,
            payload: serde_json::to_string(&payload).unwrap(),
            screenshot_path: None,
            created_at: 1000,
            retry_count,
            last_error: None,
            client_ts: 1000,
        }
    }

    // ===== 测试 =====

    #[tokio::test]
    async fn 空队列应返回0() {
        let coordinator = SyncCoordinator::new(
            Arc::new(MockReporter::always_ok()),
            Arc::new(MockBuffer::empty()),
        );
        assert_eq!(coordinator.sync_once().await, 0);
    }

    #[tokio::test]
    async fn 服务端不可用应跳过所有任务() {
        let buffer = MockBuffer::with_tasks(vec![make_valid_task(1, 0)]);
        let coordinator = SyncCoordinator::new(
            Arc::new(MockReporter::unavailable()),
            Arc::new(buffer),
        );
        assert_eq!(coordinator.sync_once().await, 0);
    }

    #[tokio::test]
    async fn 成功上报应标记完成() {
        let buffer = Arc::new(MockBuffer::with_tasks(vec![
            make_valid_task(1, 0),
            make_valid_task(2, 0),
        ]));
        let coordinator = SyncCoordinator::new(
            Arc::new(MockReporter::always_ok()),
            buffer.clone(),
        );
        let synced = coordinator.sync_once().await;
        assert_eq!(synced, 2);
        assert_eq!(buffer.completed.lock().unwrap().len(), 2);
    }

    #[tokio::test]
    async fn 上报失败应增加重试计数() {
        use crate::shared::error::AppError;
        let buffer = Arc::new(MockBuffer::with_tasks(vec![make_valid_task(1, 0)]));
        let coordinator = SyncCoordinator::new(
            Arc::new(MockReporter::with_results(vec![
                Err(AppError::Network("连接失败".to_string())),
            ])),
            buffer.clone(),
        );
        let synced = coordinator.sync_once().await;
        assert_eq!(synced, 0);
        assert_eq!(buffer.retried.lock().unwrap().len(), 1);
    }

    #[tokio::test]
    async fn 超过重试上限应直接丢弃() {
        let exhausted_task = make_valid_task(1, 20);
        let buffer = Arc::new(MockBuffer::with_tasks(vec![exhausted_task]));
        let coordinator = SyncCoordinator::new(
            Arc::new(MockReporter::always_ok()),
            buffer.clone(),
        );
        let synced = coordinator.sync_once().await;
        assert_eq!(synced, 0);
        // 超过重试次数的任务也会被标记完成（丢弃）
        assert_eq!(buffer.completed.lock().unwrap().len(), 1);
    }

    #[tokio::test]
    async fn 无效payload应标记完成并跳过() {
        let bad_task = SyncTask {
            id: 1,
            payload: "not-valid-json".to_string(),
            screenshot_path: None,
            created_at: 1000,
            retry_count: 0,
            last_error: None,
            client_ts: 1000,
        };
        let buffer = Arc::new(MockBuffer::with_tasks(vec![bad_task]));
        let coordinator = SyncCoordinator::new(
            Arc::new(MockReporter::always_ok()),
            buffer.clone(),
        );
        let synced = coordinator.sync_once().await;
        assert_eq!(synced, 0);
        assert_eq!(buffer.completed.lock().unwrap().len(), 1);
    }

    #[tokio::test]
    async fn 队列大小应正确返回() {
        let buffer = MockBuffer::with_tasks(vec![
            make_valid_task(1, 0),
            make_valid_task(2, 0),
            make_valid_task(3, 0),
        ]);
        let coordinator = SyncCoordinator::new(
            Arc::new(MockReporter::always_ok()),
            Arc::new(buffer),
        );
        assert_eq!(coordinator.queue_size().await, 3);
    }
}

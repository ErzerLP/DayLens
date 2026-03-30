//! # 活动上传器
//!
//! 将活动数据 POST 到服务端，截图通过 multipart 上传。

use std::path::Path;
use std::sync::Arc;

use async_trait::async_trait;
use tokio::sync::RwLock;

use crate::application::ports::ActivityReporter;
use crate::domain::activity::entity::Activity;
use crate::infrastructure::remote::client::RemoteClient;
use crate::shared::error::Result;

// ===== ActivityUploader =====

/// 活动上传器
pub struct ActivityUploader {
    client: Arc<RwLock<RemoteClient>>,
}

impl ActivityUploader {
    /// 创建上传器
    pub fn new(client: Arc<RwLock<RemoteClient>>) -> Self {
        Self { client }
    }
}

#[async_trait]
impl ActivityReporter for ActivityUploader {
    /// 上报活动
    ///
    /// 1. POST `/api/v1/activities` 上报活动 JSON
    /// 2. 如果有截图，POST `/api/v1/screenshots` 上传文件
    async fn report(
        &self,
        activity: &Activity,
        screenshot_path: Option<&Path>,
    ) -> Result<i64> {
        // 上报活动数据
        #[derive(serde::Deserialize)]
        struct Resp {
            id: i64,
        }

        let resp: Resp = self.client.read().await.post(
            "/api/v1/activities",
            activity,
        ).await?;

        // 上传截图（如果有）
        if let Some(path) = screenshot_path {
            if path.exists() {
                let _: serde_json::Value = self
                    .client
                    .read()
                    .await
                    .upload(
                        &format!(
                            "/api/v1/screenshots?activityId={}",
                            resp.id,
                        ),
                        path,
                        "file",
                    )
                    .await?;
            }
        }

        Ok(resp.id)
    }

    /// 批量上报
    async fn batch_report(
        &self,
        activities: &[Activity],
    ) -> Result<Vec<i64>> {
        #[derive(serde::Serialize)]
        struct Req<'a> {
            items: &'a [Activity],
        }
        #[derive(serde::Deserialize)]
        struct Resp {
            ids: Vec<i64>,
        }

        let resp: Resp = self.client.read().await.post(
            "/api/v1/activities/batch",
            &Req { items: activities },
        ).await?;

        Ok(resp.ids)
    }

    /// 健康检查
    async fn is_server_available(&self) -> bool {
        self.client.read().await.health_check().await
    }
}

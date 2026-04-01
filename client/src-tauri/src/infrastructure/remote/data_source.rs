//! # 远程数据源
//!
//! 将所有数据查询委托给 HTTP API。

use std::sync::Arc;

use async_trait::async_trait;
use tokio::sync::RwLock;

use crate::application::ports::DataSource;
use crate::domain::activity::entity::*;
use crate::infrastructure::remote::client::RemoteClient;
use crate::shared::error::Result;

// ===== RemoteDataSource =====

/// 远程数据源
pub struct RemoteDataSource {
    /// HTTP 客户端（RwLock 允许并发读、独占写配置）
    client: Arc<RwLock<RemoteClient>>,
}

impl RemoteDataSource {
    /// 创建远程数据源
    pub fn new(client: Arc<RwLock<RemoteClient>>) -> Self {
        Self { client }
    }
}

#[async_trait]
impl DataSource for RemoteDataSource {
    async fn get_stats(&self, date: &str) -> Result<DailyStats> {
        self.client.read().await.get(
            &format!("/api/v1/stats?date={date}"),
        ).await
    }

    async fn get_timeline(
        &self,
        date: &str,
        limit: i32,
        offset: i32,
        app: Option<&str>,
        category: Option<&str>,
    ) -> Result<TimelineResponse> {
        let mut url = format!(
            "/api/v1/activities?date={date}&limit={limit}&offset={offset}",
        );
        if let Some(a) = app {
            url.push_str(&format!("&app={a}"));
        }
        if let Some(c) = category {
            url.push_str(&format!("&category={c}"));
        }
        self.client.read().await.get(&url).await
    }

    async fn get_activity(&self, id: i64) -> Result<Activity> {
        self.client.read().await.get(
            &format!("/api/v1/activities/{id}"),
        ).await
    }

    async fn get_hourly_summaries(
        &self,
        date: &str,
    ) -> Result<Vec<HourlySummary>> {
        #[derive(serde::Deserialize)]
        struct Wrapper {
            #[serde(default, deserialize_with = "crate::domain::activity::entity::null_as_default")]
            items: Vec<HourlySummary>,
        }
        let wrapper: Wrapper = self.client.read().await.get(
            &format!("/api/v1/hourly-summaries?date={date}"),
        ).await?;
        Ok(wrapper.items)
    }

    async fn get_report(
        &self,
        date: &str,
    ) -> Result<Option<DailyReport>> {
        self.client.read().await.get(
            &format!("/api/v1/reports/{date}"),
        ).await
    }

    async fn generate_report(
        &self,
        date: &str,
        force_regenerate: bool,
    ) -> Result<DailyReport> {
        #[derive(serde::Serialize)]
        #[serde(rename_all = "camelCase")]
        struct Req {
            date: String,
            force_regenerate: bool,
        }
        self.client.read().await.post_long(
            "/api/v1/reports/generate",
            &Req {
                date: date.to_string(),
                force_regenerate,
            },
            120, // AI 生成可能需要较长时间
        ).await
    }

    async fn get_sessions(
        &self,
        date: &str,
    ) -> Result<Vec<WorkSession>> {
        #[derive(serde::Deserialize)]
        struct Resp {
            items: Vec<WorkSession>,
        }
        let resp: Resp = self.client.read().await.get(
            &format!("/api/v1/sessions?date={date}"),
        ).await?;
        Ok(resp.items)
    }

    async fn get_intents(
        &self,
        date: &str,
    ) -> Result<IntentAnalysisResult> {
        self.client.read().await.get(
            &format!("/api/v1/intents?date={date}"),
        ).await
    }

    async fn get_todos(
        &self,
        from: &str,
        to: &str,
    ) -> Result<TodoExtractionResult> {
        self.client.read().await.get(
            &format!("/api/v1/todos?from={from}&to={to}"),
        ).await
    }

    async fn generate_weekly_review(
        &self,
        from: &str,
        to: &str,
    ) -> Result<WeeklyReview> {
        #[derive(serde::Serialize)]
        struct Req {
            from: String,
            to: String,
        }
        self.client.read().await.post(
            "/api/v1/weekly-review",
            &Req {
                from: from.to_string(),
                to: to.to_string(),
            },
        ).await
    }

    async fn search(
        &self,
        query: &str,
        limit: i32,
    ) -> Result<Vec<SearchResultItem>> {
        #[derive(serde::Deserialize)]
        struct Resp {
            items: Vec<SearchResultItem>,
        }
        let resp: Resp = self.client.read().await.get(
            &format!("/api/v1/search?q={query}&limit={limit}"),
        ).await?;
        Ok(resp.items)
    }

    async fn ask(
        &self,
        question: &str,
        context: &str,
    ) -> Result<AiAnswer> {
        #[derive(serde::Serialize)]
        struct Req {
            question: String,
            context: String,
        }
        self.client.read().await.post(
            "/api/v1/ask",
            &Req {
                question: question.to_string(),
                context: context.to_string(),
            },
        ).await
    }

    async fn chat(
        &self,
        messages: Vec<ChatMessage>,
        tools: Vec<String>,
    ) -> Result<AssistantReply> {
        #[derive(serde::Serialize)]
        struct Req {
            messages: Vec<ChatMessage>,
            tools: Vec<String>,
        }
        self.client.read().await.post(
            "/api/v1/assistant/chat",
            &Req { messages, tools },
        ).await
    }

    async fn get_recent_apps(&self, days: i32) -> Result<Vec<String>> {
        #[derive(serde::Deserialize)]
        struct Resp {
            items: Vec<String>,
        }
        let resp: Resp = self.client.read().await.get(
            &format!("/api/v1/apps/recent?days={days}"),
        ).await?;
        Ok(resp.items)
    }

    async fn get_app_categories(
        &self,
        from: &str,
        to: &str,
    ) -> Result<Vec<AppCategoryInfo>> {
        #[derive(serde::Deserialize)]
        struct Resp {
            items: Vec<AppCategoryInfo>,
        }
        let resp: Resp = self.client.read().await.get(
            &format!("/api/v1/apps/categories?from={from}&to={to}"),
        ).await?;
        Ok(resp.items)
    }

    async fn set_category_rule(
        &self,
        app_name: &str,
        category: &str,
    ) -> Result<()> {
        #[derive(serde::Serialize)]
        #[serde(rename_all = "camelCase")]
        struct Req {
            app_name: String,
            category: String,
        }
        // 服务端返回 data: null，用 Option 包一层避免解析失败
        let _: Option<serde_json::Value> = self.client.read().await.put(
            "/api/v1/apps/category-rules",
            &Req {
                app_name: app_name.to_string(),
                category: category.to_string(),
            },
        ).await.or_else(|e| {
            // 忽略 "data 为 null" 错误，因为该接口确实不返回 data
            if format!("{e}").contains("data 为 null") {
                Ok(None)
            } else {
                Err(e)
            }
        })?;
        Ok(())
    }

    async fn reclassify_app(
        &self,
        app_name: &str,
        new_category: &str,
    ) -> Result<i64> {
        #[derive(serde::Serialize, serde::Deserialize)]
        #[serde(rename_all = "camelCase")]
        struct Req {
            app_name: String,
            new_category: String,
        }
        #[derive(serde::Deserialize)]
        #[serde(rename_all = "camelCase")]
        struct Resp {
            updated_count: i64,
        }
        let resp: Resp = self.client.read().await.post(
            "/api/v1/apps/reclassify",
            &Req {
                app_name: app_name.to_string(),
                new_category: new_category.to_string(),
            },
        ).await?;
        Ok(resp.updated_count)
    }

    async fn get_storage_stats(&self) -> Result<StorageStats> {
        self.client.read().await.get(
            "/api/v1/storage/stats",
        ).await
    }

    async fn cleanup_before(
        &self,
        date: &str,
    ) -> Result<CleanupResult> {
        self.client.read().await.delete(
            &format!("/api/v1/activities/before?date={date}"),
        ).await
    }
}

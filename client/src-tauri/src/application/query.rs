//! # 查询服务
//!
//! 将前端查询请求委托给 DataSource 端口。

use std::sync::Arc;

use crate::application::ports::DataSource;
use crate::domain::activity::entity::*;
use crate::shared::error::Result;

/// 查询服务 — 前端查询的统一入口
pub struct QueryService {
    data_source: Arc<dyn DataSource>,
}

impl QueryService {
    pub fn new(data_source: Arc<dyn DataSource>) -> Self {
        Self { data_source }
    }

    pub async fn get_stats(&self, date: &str) -> Result<DailyStats> {
        self.data_source.get_stats(date).await
    }

    pub async fn get_timeline(
        &self,
        date: &str,
        limit: i32,
        offset: i32,
        app: Option<&str>,
        category: Option<&str>,
    ) -> Result<TimelineResponse> {
        self.data_source
            .get_timeline(date, limit, offset, app, category)
            .await
    }

    pub async fn get_activity(&self, id: i64) -> Result<Activity> {
        self.data_source.get_activity(id).await
    }

    pub async fn get_hourly_summaries(
        &self,
        date: &str,
    ) -> Result<Vec<HourlySummary>> {
        self.data_source.get_hourly_summaries(date).await
    }

    pub async fn get_report(
        &self,
        date: &str,
    ) -> Result<Option<DailyReport>> {
        self.data_source.get_report(date).await
    }

    pub async fn generate_report(
        &self,
        date: &str,
        force: bool,
    ) -> Result<DailyReport> {
        self.data_source.generate_report(date, force).await
    }

    pub async fn get_sessions(
        &self,
        date: &str,
    ) -> Result<Vec<WorkSession>> {
        self.data_source.get_sessions(date).await
    }

    pub async fn search(
        &self,
        query: &str,
        limit: i32,
    ) -> Result<Vec<SearchResultItem>> {
        self.data_source.search(query, limit).await
    }

    pub async fn ask(
        &self,
        question: &str,
        context: &str,
    ) -> Result<AiAnswer> {
        self.data_source.ask(question, context).await
    }

    pub async fn chat(
        &self,
        messages: Vec<ChatMessage>,
        tools: Vec<String>,
    ) -> Result<AssistantReply> {
        self.data_source.chat(messages, tools).await
    }

    pub async fn get_storage_stats(&self) -> Result<StorageStats> {
        self.data_source.get_storage_stats().await
    }

    pub async fn cleanup_before(
        &self,
        date: &str,
    ) -> Result<CleanupResult> {
        self.data_source.cleanup_before(date).await
    }

    pub async fn get_app_categories(
        &self,
        from: &str,
        to: &str,
    ) -> Result<Vec<AppCategoryInfo>> {
        self.data_source.get_app_categories(from, to).await
    }

    pub async fn set_category_rule(
        &self,
        app_name: &str,
        category: &str,
    ) -> Result<()> {
        self.data_source.set_category_rule(app_name, category).await
    }

    pub async fn reclassify_app(
        &self,
        app_name: &str,
        new_category: &str,
    ) -> Result<i64> {
        self.data_source.reclassify_app(app_name, new_category).await
    }
}

//! # 应用层端口定义
//!
//! 所有端口（trait）定义在此文件中。
//! 基础设施层实现这些 trait，应用层通过 `Arc<dyn Trait>` 调用。
//! 端口签名中的参数和返回值 **只允许使用领域层类型**。

use std::path::Path;

use async_trait::async_trait;

use crate::domain::activity::entity::{
    Activity, AiAnswer, AppCategoryInfo, AssistantReply,
    ChatMessage, CleanupResult, DailyReport, DailyStats, HourlySummary,
    IntentAnalysisResult, SearchResultItem, StorageStats,
    TimelineResponse, TodoExtractionResult, WeeklyReview, WorkSession,
    ActiveWindow,
};
use crate::domain::sync::entity::SyncTask;
use crate::shared::error::Result;

// ===== 数据查询端口 =====

/// 数据查询端口
#[async_trait]
pub trait DataSource: Send + Sync {
    /// 获取指定日期的统计数据
    async fn get_stats(&self, date: &str) -> Result<DailyStats>;

    /// 获取活动时间线（分页）
    async fn get_timeline(
        &self,
        date: &str,
        limit: i32,
        offset: i32,
        app: Option<&str>,
        category: Option<&str>,
    ) -> Result<TimelineResponse>;

    /// 获取单条活动详情
    async fn get_activity(&self, id: i64) -> Result<Activity>;

    /// 获取小时摘要列表
    async fn get_hourly_summaries(
        &self,
        date: &str,
    ) -> Result<Vec<HourlySummary>>;

    /// 获取日报
    async fn get_report(&self, date: &str) -> Result<Option<DailyReport>>;

    /// 触发生成日报
    async fn generate_report(
        &self,
        date: &str,
        force_regenerate: bool,
    ) -> Result<DailyReport>;

    /// 获取工作会话列表
    async fn get_sessions(&self, date: &str) -> Result<Vec<WorkSession>>;

    /// 获取意图分析结果
    async fn get_intents(
        &self,
        date: &str,
    ) -> Result<IntentAnalysisResult>;

    /// 获取待办事项
    async fn get_todos(
        &self,
        from: &str,
        to: &str,
    ) -> Result<TodoExtractionResult>;

    /// 生成周报
    async fn generate_weekly_review(
        &self,
        from: &str,
        to: &str,
    ) -> Result<WeeklyReview>;

    /// 全文搜索
    async fn search(
        &self,
        query: &str,
        limit: i32,
    ) -> Result<Vec<SearchResultItem>>;

    /// AI 问答
    async fn ask(
        &self,
        question: &str,
        context: &str,
    ) -> Result<AiAnswer>;

    /// AI 助手对话
    async fn chat(
        &self,
        messages: Vec<ChatMessage>,
        tools: Vec<String>,
    ) -> Result<AssistantReply>;

    /// 获取最近使用的应用列表
    async fn get_recent_apps(&self, days: i32) -> Result<Vec<String>>;

    /// 获取应用分类概览
    async fn get_app_categories(
        &self,
        from: &str,
        to: &str,
    ) -> Result<Vec<AppCategoryInfo>>;

    /// 设置应用分类规则
    async fn set_category_rule(
        &self,
        app_name: &str,
        category: &str,
    ) -> Result<()>;

    /// 重新分类应用历史
    async fn reclassify_app(
        &self,
        app_name: &str,
        new_category: &str,
    ) -> Result<i64>;

    /// 获取存储统计
    async fn get_storage_stats(&self) -> Result<StorageStats>;

    /// 清理旧数据
    async fn cleanup_before(
        &self,
        date: &str,
    ) -> Result<CleanupResult>;
}

/// 活动上报端口
#[async_trait]
pub trait ActivityReporter: Send + Sync {
    /// 上报单条活动 + 可选截图
    async fn report(
        &self,
        activity: &Activity,
        screenshot_path: Option<&Path>,
    ) -> Result<i64>;

    /// 批量上报活动
    async fn batch_report(
        &self,
        activities: &[Activity],
    ) -> Result<Vec<i64>>;

    /// 服务端可用性检查
    async fn is_server_available(&self) -> bool;
}

// ===== 离线缓冲端口 =====

/// 离线缓冲端口
pub trait SyncBuffer: Send + Sync {
    /// 将任务加入缓冲队列
    fn enqueue(&self, task: &SyncTask) -> Result<()>;

    /// 获取待重试任务
    fn pending_tasks(&self, limit: i32) -> Result<Vec<SyncTask>>;

    /// 标记任务完成（删除）
    fn mark_completed(&self, task_id: &str) -> Result<()>;

    /// 增加重试计数
    fn increment_retry(&self, task_id: &str) -> Result<()>;

    /// 获取队列长度
    fn queue_size(&self) -> Result<i64>;
}

// ===== 事件发射端口 =====

/// 前端事件推送端口
pub trait EventEmitter: Send + Sync {
    fn emit(&self, event: &str, payload: &str) -> Result<()>;
}

// ===== 截屏端口 =====

/// 截屏端口
#[async_trait]
pub trait ScreenCapture: Send + Sync {
    /// 截取当前屏幕并保存到指定路径
    async fn capture(&self, save_path: &Path) -> Result<()>;

    fn generate_thumbnail(
        &self,
        source: &Path,
        target: &Path,
        width: u32,
    ) -> Result<()>;
}

// ===== OCR 端口 =====

/// OCR 文字识别端口
#[async_trait]
pub trait OcrEngine: Send + Sync {
    async fn recognize(&self, image_path: &Path) -> Result<String>;

    fn is_available(&self) -> bool;
}

// ===== 窗口监控端口 =====

/// 窗口监控端口
pub trait WindowMonitor: Send + Sync {
    fn get_active_window(&self) -> Result<ActiveWindow>;

    fn get_browser_url(&self, app_name: &str) -> Option<String>;
}

// ===== 空闲检测端口 =====

/// 空闲检测端口
pub trait IdleDetector: Send + Sync {
    fn is_idle(&self) -> bool;
    fn reset(&self);
}

// ===== 锁屏检测端口 =====

/// 锁屏检测端口
pub trait ScreenLockDetector: Send + Sync {
    fn is_locked(&self) -> bool;
}

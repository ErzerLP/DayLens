//! # 活动领域实体
//!
//! 定义活动采集、统计、应用使用等核心数据结构。
//! 所有结构体均为纯数据，字段 `pub`，可跨层使用。

use serde::{Deserialize, Deserializer, Serialize};

/// 遇到 null 时返回类型默认值（用于 Vec 等字段）
fn null_as_default<'de, D, T>(deserializer: D) -> std::result::Result<T, D::Error>
where
    D: Deserializer<'de>,
    T: Default + Deserialize<'de>,
{
    Ok(Option::<T>::deserialize(deserializer)?.unwrap_or_default())
}

// ===== 核心实体 =====

/// 一条活动记录 — 系统采集的单次窗口活动快照
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "camelCase", default)]
pub struct Activity {
    /// 服务端分配的 ID（上报前为 0）
    pub id: i64,
    /// 客户端设备唯一标识
    pub client_id: String,
    /// 客户端采集时间戳（秒级，幂等键）
    pub client_ts: i64,
    /// 活动发生时间戳（秒级）
    pub timestamp: i64,
    /// 应用程序名（归一化后）
    pub app_name: String,
    /// 窗口标题（脱敏后可能为 `[内容已脱敏]`）
    pub window_title: String,
    /// 基础分类
    pub category: String,
    /// 语义分类
    pub semantic_category: Option<String>,
    /// 语义分类置信度 0-100
    pub semantic_confidence: Option<i32>,
    /// 持续时长（秒）
    pub duration: i32,
    /// 浏览器 URL（脱敏后可能为空）
    pub browser_url: Option<String>,
    /// 可执行文件路径
    pub executable_path: Option<String>,
    /// OCR 识别文本
    pub ocr_text: Option<String>,
    /// 截图存储键
    pub screenshot_key: Option<String>,
}

/// 每日统计 — 某天的所有汇总数据
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "camelCase", default)]
pub struct DailyStats {
    /// 日期 YYYY-MM-DD
    pub date: String,
    /// 总活跃时长（秒）
    pub total_duration: i64,
    /// 截图总数
    pub screenshot_count: i64,
    /// 活跃小时数
    pub active_hours: i32,
    /// 应用使用分布
    #[serde(deserialize_with = "null_as_default")]
    pub app_usage: Vec<AppUsage>,
    /// 分类使用分布
    #[serde(deserialize_with = "null_as_default")]
    pub category_usage: Vec<CategoryUsage>,
    /// 域名使用分布
    #[serde(deserialize_with = "null_as_default")]
    pub domain_usage: Vec<DomainUsage>,
    /// 工作时间段内的总时长（秒）
    pub work_time_duration: i64,
    /// 最常用窗口标题
    #[serde(deserialize_with = "null_as_default")]
    pub top_window_titles: Vec<WindowTitleUsage>,
}

/// 应用使用统计
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AppUsage {
    /// 应用名
    pub app_name: String,
    /// 使用总时长（秒）
    pub duration: i64,
    /// 活动记录数
    pub count: i64,
}

/// 分类使用统计
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CategoryUsage {
    /// 分类名
    pub category: String,
    /// 使用总时长（秒）
    pub duration: i64,
}

/// 域名使用统计
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DomainUsage {
    /// 域名
    pub domain: String,
    /// 使用总时长（秒）
    pub duration: i64,
}

/// 窗口标题使用统计
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct WindowTitleUsage {
    /// 窗口标题
    pub title: String,
    /// 使用总时长（秒）
    pub duration: i64,
    /// 所属应用名
    pub app_name: String,
}

/// 小时摘要
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
#[serde(rename_all = "camelCase", default)]
pub struct HourlySummary {
    /// 小时（0-23）
    pub hour: i32,
    /// 该小时的概括描述
    pub summary: String,
    /// 主要使用的应用
    pub main_apps: String,
    /// 活动记录数
    pub activity_count: i64,
    /// 总活跃时长（秒）
    pub total_duration: i64,
    /// 代表性截图键列表
    #[serde(deserialize_with = "null_as_default")]
    pub representative_screenshots: Vec<String>,
}

/// 活动列表分页响应
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TimelineResponse {
    /// 活动列表
    pub items: Vec<Activity>,
    /// 总记录数
    pub total: i64,
    /// 每页数量
    pub limit: i32,
    /// 偏移量
    pub offset: i32,
}

// ===== 日报 =====

/// 日报
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DailyReport {
    /// 日期 YYYY-MM-DD
    pub date: String,
    /// Markdown 内容
    pub content: String,
    /// AI 模式
    pub ai_mode: String,
    /// 模型名
    pub model_name: String,
    /// 是否使用了 AI
    pub used_ai: bool,
    /// 创建时间戳
    pub created_at: i64,
    /// 更新时间戳
    pub updated_at: i64,
}

// ===== 工作智能 =====

/// 工作会话
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct WorkSession {
    /// 会话开始时间戳
    pub start_time: i64,
    /// 会话结束时间戳
    pub end_time: i64,
    /// 总时长（秒）
    pub total_duration: i64,
    /// 会话内的活动列表
    pub activities: Vec<SessionActivity>,
    /// 主导应用
    pub dominant_app: String,
    /// 意图分析
    pub intent: IntentInfo,
}

/// 会话内的活动摘要
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SessionActivity {
    /// 应用名
    pub app_name: String,
    /// 使用时长（秒）
    pub duration: i64,
    /// 窗口标题
    pub title: String,
}

/// 意图信息
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct IntentInfo {
    /// 意图标签
    pub label: String,
    /// 置信度 0-100
    pub confidence: i32,
    /// 判断依据
    pub evidence: Vec<String>,
}

/// 意图分析条目
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct IntentItem {
    /// 意图标签
    pub label: String,
    /// 总时长（秒）
    pub total_duration: i64,
    /// 会话数
    pub session_count: i32,
    /// 占比 0-100
    pub percentage: i32,
}

/// 意图分析结果
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct IntentAnalysisResult {
    /// 各意图统计
    pub items: Vec<IntentItem>,
    /// 主导意图
    pub dominant_intent: String,
    /// 分析覆盖的总时长（秒）
    pub total_analyzed_duration: i64,
}

/// 待办事项
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TodoItem {
    /// 待办标题
    pub title: String,
    /// 来源
    pub source: String,
    /// 置信度等级
    pub confidence: String,
    /// 提取时间戳
    pub extracted_at: i64,
}

/// 待办提取结果
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct TodoExtractionResult {
    /// 待办列表
    pub items: Vec<TodoItem>,
    /// 汇总描述
    pub summary: String,
}

/// 周报
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct WeeklyReview {
    /// 时间段描述
    pub period: String,
    /// 总工作时长（秒）
    pub total_work_duration: i64,
    /// 日均工作时长（秒）
    pub avg_daily_duration: i64,
    /// Markdown 内容
    pub content: String,
    /// 深度工作会话
    pub deep_work_sessions: Vec<DeepWorkSession>,
    /// 最常用应用
    pub top_apps: Vec<AppUsage>,
    /// 意图分布
    pub intent_distribution: Vec<IntentItem>,
}

/// 深度工作会话
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct DeepWorkSession {
    /// 日期
    pub date: String,
    /// 时长（秒）
    pub duration: i64,
    /// 专注领域
    pub focus: String,
}

// ===== 搜索与 AI =====

/// 搜索结果条目
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct SearchResultItem {
    /// 关联活动 ID
    pub activity_id: i64,
    /// 时间戳
    pub timestamp: i64,
    /// 应用名
    pub app_name: String,
    /// 匹配片段
    pub excerpt: String,
    /// 匹配字段名
    pub match_field: String,
    /// 相关性分数
    pub relevance_score: i64,
}

/// AI 问答回答
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AiAnswer {
    /// 回答内容
    pub answer: String,
    /// 引用的活动记录
    pub references: Vec<AiReference>,
}

/// AI 回答引用
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AiReference {
    /// 活动 ID
    pub activity_id: i64,
    /// 时间戳
    pub timestamp: i64,
    /// 摘录
    pub excerpt: String,
}

/// AI 助手对话消息
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChatMessage {
    /// 角色：user / assistant / system
    pub role: String,
    /// 消息内容
    pub content: String,
}

/// AI 助手回复
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AssistantReply {
    /// 回复内容
    pub reply: String,
    /// 工具调用记录
    pub tool_calls: Vec<ToolCall>,
}

/// AI 工具调用
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ToolCall {
    /// 工具名
    pub tool: String,
    /// 输入参数 JSON
    pub input: serde_json::Value,
    /// 输出结果
    pub output: String,
}

// ===== 应用管理 =====

/// 应用分类信息
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AppCategoryInfo {
    /// 应用名
    pub app_name: String,
    /// 分类
    pub category: String,
    /// 是否为自定义规则
    pub is_custom_rule: bool,
    /// 总使用时长（秒）
    pub total_duration: i64,
    /// 最后使用时间戳
    pub last_seen: i64,
}

// ===== 存储管理 =====

/// 存储统计
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct StorageStats {
    /// 活动记录总数
    pub activity_count: i64,
    /// 截图总数
    pub screenshot_count: i64,
    /// 磁盘使用量（MB）
    pub disk_usage_mb: i64,
    /// 最大存储量（MB）
    pub max_storage_mb: i64,
    /// 最早活动日期
    pub oldest_activity_date: String,
    /// 保留天数
    pub retention_days: i32,
}

/// 数据清理结果
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CleanupResult {
    /// 删除的活动数
    pub deleted_activities: i64,
    /// 删除的截图数
    pub deleted_screenshots: i64,
    /// 释放的空间（MB）
    pub freed_mb: i64,
}

// ===== 批量上报 =====

/// 批量上报结果
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct BatchResult {
    /// 总提交数
    pub total: i32,
    /// 新插入数
    pub inserted: i32,
    /// 幂等命中（重复跳过）数
    pub deduplicated: i32,
}

// ===== 窗口监控 =====

/// 前台窗口信息 — 由 WindowMonitor 端口返回
#[derive(Debug, Clone)]
pub struct ActiveWindow {
    /// 应用程序名（归一化后，如 "Code"）
    pub app_name: String,
    /// 窗口标题
    pub window_title: String,
    /// 可执行文件完整路径
    pub executable_path: String,
    /// 是否为浏览器应用
    pub is_browser: bool,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 活动实体可序列化为json() {
        let activity = Activity {
            id: 0,
            client_id: "test-device".to_string(),
            client_ts: 1711699200,
            timestamp: 1711699200,
            app_name: "Code".to_string(),
            window_title: "main.rs".to_string(),
            category: "coding".to_string(),
            semantic_category: Some("Coding".to_string()),
            semantic_confidence: Some(92),
            duration: 30,
            browser_url: None,
            executable_path: None,
            ocr_text: None,
            screenshot_key: None,
        };

        let json = serde_json::to_string(&activity)
            .expect("序列化失败");
        assert!(json.contains("\"appName\":\"Code\""));
        assert!(json.contains("\"clientTs\":1711699200"));
    }

    #[test]
    fn 每日统计可从json反序列化() {
        let json = r#"{
            "date": "2026-03-29",
            "totalDuration": 28800,
            "screenshotCount": 960,
            "activeHours": 8,
            "appUsage": [],
            "categoryUsage": [],
            "domainUsage": [],
            "workTimeDuration": 25200,
            "topWindowTitles": []
        }"#;

        let stats: DailyStats = serde_json::from_str(json)
            .expect("反序列化失败");
        assert_eq!(stats.total_duration, 28800);
        assert_eq!(stats.active_hours, 8);
    }

    #[test]
    fn 批量结果应正确反序列化() {
        let json = r#"{"total": 10, "inserted": 8, "deduplicated": 2}"#;
        let result: BatchResult = serde_json::from_str(json)
            .expect("反序列化失败");
        assert_eq!(result.total, 10);
        assert_eq!(result.inserted, 8);
        assert_eq!(result.deduplicated, 2);
    }
}

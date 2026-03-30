//! # 配置领域实体
//!
//! 定义应用所有配置数据结构：服务端连接、隐私规则、工作时间等。

use serde::{Deserialize, Serialize};

use crate::domain::activity::value_objects::PrivacyAction;

/// 应用全局配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    /// 服务端连接配置
    pub server: ServerConfig,
    /// 隐私规则列表
    pub privacy_rules: Vec<PrivacyRule>,
    /// 采集配置
    pub capture: CaptureConfig,
    /// 工作时间配置
    pub work_schedule: WorkSchedule,
    /// AI 配置
    pub ai: AiConfig,
}

/// 服务端连接配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerConfig {
    /// 服务端 URL（如 `https://your-server.com:8080`）
    pub url: String,
    /// 认证 Token
    pub token: String,
    /// 是否上传截图文件
    pub upload_screenshots: bool,
    /// 启用离线缓冲
    pub offline_buffer: bool,
    /// 缓冲同步间隔（秒）
    pub sync_interval_secs: u64,
}

/// 隐私规则
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PrivacyRule {
    /// 应用名匹配模式（包含匹配）
    pub app_name_pattern: Option<String>,
    /// 标题敏感关键词列表
    pub title_keywords: Option<Vec<String>>,
    /// 域名匹配模式列表
    pub domain_patterns: Option<Vec<String>>,
    /// 匹配后执行的动作
    pub action: PrivacyAction,
}

/// 采集配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CaptureConfig {
    /// 截屏间隔（秒）
    pub screenshot_interval_secs: u64,
    /// 空闲超时（分钟）— 超过此时间无输入视为空闲
    pub idle_timeout_minutes: u32,
    /// 是否启用 OCR
    pub enable_ocr: bool,
    /// 缩略图宽度（像素）
    pub thumbnail_width: u32,
}

/// 工作时间配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkSchedule {
    /// 是否启用工作时间限制
    pub enabled: bool,
    /// 开始时间（如 "09:00"）
    pub start_time: String,
    /// 结束时间（如 "18:00"）
    pub end_time: String,
    /// 工作日列表（1=周一, 7=周日）
    pub work_days: Vec<u8>,
}

impl WorkSchedule {
    /// 判断当前时刻是否在工作时间范围内
    ///
    /// 规则：未启用时始终返回 true；启用时检查当前星期 + 时段。
    pub fn is_work_time(&self) -> bool {
        if !self.enabled {
            return true;
        }

        let now = chrono::Local::now();
        use chrono::Datelike;
        // chrono weekday: Mon=0 ... Sun=6 → 转换为 1..=7
        let weekday = now.weekday().num_days_from_monday() as u8 + 1;
        if !self.work_days.contains(&weekday) {
            return false;
        }

        let time_str = now.format("%H:%M").to_string();
        time_str >= self.start_time && time_str <= self.end_time
    }
}

/// AI 提供商配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AiConfig {
    /// 提供商 ID
    pub provider: String,
    /// API 端点
    pub endpoint: String,
    /// 模型名
    pub model: String,
    /// API Key
    pub api_key: String,
    /// 自定义 System Prompt
    pub custom_prompt: String,
}

// ===== PrivacyAction 序列化支持 =====

impl Serialize for PrivacyAction {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        match self {
            Self::Allow => serializer.serialize_str("allow"),
            Self::Anonymize => serializer.serialize_str("anonymize"),
            Self::Skip => serializer.serialize_str("skip"),
        }
    }
}

impl<'de> Deserialize<'de> for PrivacyAction {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        let s = String::deserialize(deserializer)?;
        match s.as_str() {
            "allow" => Ok(Self::Allow),
            "anonymize" => Ok(Self::Anonymize),
            "skip" => Ok(Self::Skip),
            other => Err(serde::de::Error::custom(
                format!("未知的隐私动作: {other}"),
            )),
        }
    }
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            server: ServerConfig {
                url: "http://localhost:8080".to_string(),
                token: String::new(),
                upload_screenshots: true,
                offline_buffer: true,
                sync_interval_secs: 60,
            },
            privacy_rules: Vec::new(),
            capture: CaptureConfig {
                screenshot_interval_secs: 30,
                idle_timeout_minutes: 3,
                enable_ocr: true,
                thumbnail_width: 360,
            },
            work_schedule: WorkSchedule {
                enabled: false,
                start_time: "09:00".to_string(),
                end_time: "18:00".to_string(),
                work_days: vec![1, 2, 3, 4, 5],
            },
            ai: AiConfig {
                provider: "ollama".to_string(),
                endpoint: "http://localhost:11434".to_string(),
                model: "qwen2.5".to_string(),
                api_key: String::new(),
                custom_prompt: String::new(),
            },
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 默认配置应有合理的初始值() {
        let config = AppConfig::default();
        assert_eq!(config.server.url, "http://localhost:8080");
        assert!(config.server.upload_screenshots);
        assert_eq!(config.capture.screenshot_interval_secs, 30);
        assert_eq!(config.capture.idle_timeout_minutes, 3);
        assert_eq!(config.work_schedule.work_days.len(), 5);
    }

    #[test]
    fn 配置可序列化和反序列化() {
        let config = AppConfig::default();
        let json = serde_json::to_string(&config)
            .expect("序列化失败");
        let restored: AppConfig = serde_json::from_str(&json)
            .expect("反序列化失败");
        assert_eq!(restored.server.url, config.server.url);
    }

    #[test]
    fn 隐私动作可序列化() {
        let rule = PrivacyRule {
            app_name_pattern: Some("wechat".to_string()),
            title_keywords: None,
            domain_patterns: None,
            action: PrivacyAction::Skip,
        };
        let json = serde_json::to_string(&rule)
            .expect("序列化失败");
        assert!(json.contains("\"skip\""));
    }

    #[test]
    fn 隐私动作可反序列化() {
        let json = r#"{
            "appNamePattern": null,
            "titleKeywords": null,
            "domainPatterns": null,
            "action": "anonymize"
        }"#;
        let rule: PrivacyRule = serde_json::from_str(json)
            .expect("反序列化失败");
        assert_eq!(rule.action, PrivacyAction::Anonymize);
    }

    #[test]
    fn 工作时间_未启用应始终返回true() {
        let schedule = WorkSchedule {
            enabled: false,
            start_time: "09:00".to_string(),
            end_time: "18:00".to_string(),
            work_days: vec![1, 2, 3, 4, 5],
        };
        assert!(schedule.is_work_time());
    }

    #[test]
    fn 工作时间_全天候应返回true() {
        let schedule = WorkSchedule {
            enabled: true,
            start_time: "00:00".to_string(),
            end_time: "23:59".to_string(),
            work_days: vec![1, 2, 3, 4, 5, 6, 7],
        };
        assert!(schedule.is_work_time());
    }
}

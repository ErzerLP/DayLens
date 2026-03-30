//! # 活动领域值对象
//!
//! 定义不可变的判定结果类型：隐私动作、基础分类、语义分类。
//! 值对象没有身份标识，仅通过值来区分。

use serde::{Deserialize, Serialize};

// ===== 隐私动作 =====

/// 隐私规则判定后的动作
///
/// 决定一条活动数据在离开本机前应如何处理。
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum PrivacyAction {
    /// 允许上报原始数据
    Allow,
    /// 脱敏处理（替换标题/URL/OCR 文本为 `[内容已脱敏]`）
    Anonymize,
    /// 完全跳过，不记录此活动
    Skip,
}

// ===== 基础分类 =====

/// 基础分类 — 根据应用名进行粗粒度分类
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum Category {
    /// 编码开发
    #[serde(rename = "coding")]
    Coding,
    /// 网页浏览
    #[serde(rename = "browser")]
    Browser,
    /// 即时通讯
    #[serde(rename = "communication")]
    Communication,
    /// 文档编辑
    #[serde(rename = "document")]
    Document,
    /// 设计工具
    #[serde(rename = "design")]
    Design,
    /// 终端命令行
    #[serde(rename = "terminal")]
    Terminal,
    /// 媒体播放
    #[serde(rename = "media")]
    Media,
    /// 系统工具
    #[serde(rename = "system")]
    System,
    /// 游戏娱乐
    #[serde(rename = "gaming")]
    Gaming,
    /// 其他
    #[serde(rename = "other")]
    Other,
}

impl Category {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Coding => "coding",
            Self::Browser => "browser",
            Self::Communication => "communication",
            Self::Document => "document",
            Self::Design => "design",
            Self::Terminal => "terminal",
            Self::Media => "media",
            Self::System => "system",
            Self::Gaming => "gaming",
            Self::Other => "other",
        }
    }
}

// ===== 语义分类 =====

/// 语义分类 — 结合标题/URL/OCR 进行细粒度分类
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum SemanticCategory {
    /// 编码开发
    Coding,
    /// 资料调研
    Research,
    /// 沟通协作
    Communication,
    /// 文档写作
    Writing,
    /// 设计创作
    Design,
    /// 一般浏览
    Browsing,
    /// 娱乐放松
    Entertainment,
    /// 系统操作
    SystemOperation,
    /// 其他
    Other,
}

impl SemanticCategory {
    /// 将枚举转为 API 对应的字符串值
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Coding => "Coding",
            Self::Research => "Research",
            Self::Communication => "Communication",
            Self::Writing => "Writing",
            Self::Design => "Design",
            Self::Browsing => "Browsing",
            Self::Entertainment => "Entertainment",
            Self::SystemOperation => "SystemOperation",
            Self::Other => "Other",
        }
    }
}

/// 分类结果 — 由 ClassifierService 返回
#[derive(Debug, Clone)]
pub struct ClassificationResult {
    /// 基础分类
    pub category: Category,
    /// 语义分类
    pub semantic_category: SemanticCategory,
    /// 语义分类置信度 0-100
    pub confidence: i32,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 隐私动作值相等性比较() {
        assert_eq!(PrivacyAction::Allow, PrivacyAction::Allow);
        assert_ne!(PrivacyAction::Allow, PrivacyAction::Skip);
    }

    #[test]
    fn 基础分类可转为字符串() {
        assert_eq!(Category::Coding.as_str(), "coding");
        assert_eq!(Category::Browser.as_str(), "browser");
        assert_eq!(Category::Terminal.as_str(), "terminal");
    }

    #[test]
    fn 语义分类可转为字符串() {
        assert_eq!(SemanticCategory::Coding.as_str(), "Coding");
        assert_eq!(SemanticCategory::Research.as_str(), "Research");
    }

    #[test]
    fn 基础分类可序列化为json() {
        let json = serde_json::to_string(&Category::Coding)
            .expect("序列化失败");
        assert_eq!(json, "\"coding\"");
    }

    #[test]
    fn 基础分类可从json反序列化() {
        let category: Category = serde_json::from_str("\"browser\"")
            .expect("反序列化失败");
        assert_eq!(category, Category::Browser);
    }
}

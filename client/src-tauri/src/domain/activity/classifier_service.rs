//! # 活动分类领域服务
//!
//! 纯同步业务规则：根据应用名、窗口标题、浏览器 URL 判定活动类别。
//! 先通过应用名映射表确定基础分类，再结合标题/URL 推断语义分类。

use crate::domain::activity::value_objects::{
    Category, ClassificationResult, SemanticCategory,
};

/// 活动分类领域服务
///
/// 无状态服务，所有方法为关联函数。
pub struct ClassifierService;

impl ClassifierService {
    /// 对一条活动进行分类
    ///
    /// # 参数
    /// - `app_name` — 归一化后的应用名
    /// - `window_title` — 窗口标题
    /// - `browser_url` — 浏览器 URL（可选）
    ///
    /// # 返回
    /// 包含基础分类和语义分类的结果
    pub fn classify(
        app_name: &str,
        window_title: &str,
        browser_url: Option<&str>,
    ) -> ClassificationResult {
        let category = Self::classify_by_app(app_name);
        let (semantic, confidence) = Self::classify_semantic(
            &category,
            app_name,
            window_title,
            browser_url,
        );

        ClassificationResult {
            category,
            semantic_category: semantic,
            confidence,
        }
    }

    /// 根据应用名确定基础分类
    fn classify_by_app(app_name: &str) -> Category {
        let name = app_name.to_lowercase();

        // 编码开发
        if matches!(
            name.as_str(),
            "code" | "cursor" | "idea" | "intellij"
            | "pycharm" | "webstorm" | "goland" | "clion"
            | "rider" | "rustrover" | "datagrip"
            | "sublime_text" | "sublime text"
            | "notepad++" | "vim" | "neovim" | "nvim"
            | "emacs" | "helix" | "zed" | "fleet"
            | "android studio" | "xcode"
        ) {
            return Category::Coding;
        }

        // 浏览器
        if matches!(
            name.as_str(),
            "chrome" | "msedge" | "edge" | "firefox"
            | "brave" | "opera" | "vivaldi" | "arc"
            | "safari" | "chromium"
        ) {
            return Category::Browser;
        }

        // 即时通讯
        if matches!(
            name.as_str(),
            "wechat" | "微信" | "qq" | "dingtalk" | "钉钉"
            | "feishu" | "飞书" | "lark" | "slack" | "teams"
            | "discord" | "telegram" | "zoom" | "tencent meeting"
        ) {
            return Category::Communication;
        }

        // 文档编辑
        if matches!(
            name.as_str(),
            "word" | "winword" | "excel" | "powerpoint"
            | "powerpnt" | "onenote" | "notion" | "obsidian"
            | "typora" | "wps" | "wps office"
            | "pages" | "numbers" | "keynote"
        ) {
            return Category::Document;
        }

        // 终端
        if matches!(
            name.as_str(),
            "windowsterminal" | "windows terminal"
            | "cmd" | "powershell" | "pwsh"
            | "warp" | "alacritty" | "wezterm" | "kitty"
            | "iterm2" | "terminal" | "hyper"
        ) {
            return Category::Terminal;
        }

        // 设计
        if matches!(
            name.as_str(),
            "figma" | "sketch" | "photoshop" | "illustrator"
            | "canva" | "affinity" | "inkscape" | "gimp"
            | "blender" | "after effects"
        ) {
            return Category::Design;
        }

        // 媒体
        if matches!(
            name.as_str(),
            "spotify" | "vlc" | "网易云音乐" | "cloudmusic"
            | "qqmusic" | "potplayer" | "mpv"
            | "foobar2000" | "musicbee"
        ) {
            return Category::Media;
        }

        // 系统工具
        if matches!(
            name.as_str(),
            "explorer" | "finder" | "task manager"
            | "taskmgr" | "devenv" | "regedit"
            | "everything" | "7zip" | "winrar"
        ) {
            return Category::System;
        }

        // 游戏
        if matches!(
            name.as_str(),
            "steam" | "epic games" | "origin"
            | "battle.net" | "genshin impact"
        ) {
            return Category::Gaming;
        }

        Category::Other
    }

    /// 结合标题/URL 推断语义分类
    fn classify_semantic(
        base_category: &Category,
        _app_name: &str,
        window_title: &str,
        browser_url: Option<&str>,
    ) -> (SemanticCategory, i32) {
        let title_lower = window_title.to_lowercase();

        // 浏览器需要根据 URL/标题进一步分类
        if *base_category == Category::Browser {
            return Self::classify_browser_semantic(
                &title_lower,
                browser_url,
            );
        }

        // 其他应用直接从基础分类映射
        let semantic = match base_category {
            Category::Coding => SemanticCategory::Coding,
            Category::Terminal => SemanticCategory::Coding,
            Category::Communication => SemanticCategory::Communication,
            Category::Document => {
                if Self::has_any_keyword(&title_lower, &[
                    "报告", "文档", "笔记", "总结", "方案",
                    "report", "doc", "note", "summary",
                ]) {
                    SemanticCategory::Writing
                } else {
                    SemanticCategory::Writing
                }
            }
            Category::Design => SemanticCategory::Design,
            Category::Media => SemanticCategory::Entertainment,
            Category::Gaming => SemanticCategory::Entertainment,
            Category::System => SemanticCategory::SystemOperation,
            _ => SemanticCategory::Other,
        };

        // 基于确定性的应用映射，置信度较高
        (semantic, 85)
    }

    /// 浏览器语义分类 — 根据 URL 和标题进一步细分
    fn classify_browser_semantic(
        title_lower: &str,
        browser_url: Option<&str>,
    ) -> (SemanticCategory, i32) {
        let url_lower = browser_url
            .unwrap_or("")
            .to_lowercase();

        // 编码相关网站
        if Self::has_any_keyword(&url_lower, &[
            "github.com", "gitlab.com", "stackoverflow.com",
            "stackexchange.com", "crates.io", "docs.rs",
            "npmjs.com", "pypi.org", "pkg.go.dev",
        ]) {
            return (SemanticCategory::Coding, 90);
        }

        // 调研/学习
        if Self::has_any_keyword(&url_lower, &[
            "wikipedia.org", "zhihu.com", "juejin.cn",
            "csdn.net", "cnblogs.com", "segmentfault.com",
            "medium.com", "dev.to", "arxiv.org",
        ]) || Self::has_any_keyword(title_lower, &[
            "教程", "文档", "tutorial", "guide", "reference",
        ]) {
            return (SemanticCategory::Research, 80);
        }

        // 沟通
        if Self::has_any_keyword(&url_lower, &[
            "mail.google.com", "outlook.live.com",
            "mail.qq.com", "web.telegram.org",
            "web.whatsapp.com",
        ]) {
            return (SemanticCategory::Communication, 85);
        }

        // 娱乐
        if Self::has_any_keyword(&url_lower, &[
            "youtube.com", "bilibili.com", "netflix.com",
            "twitch.tv", "v.qq.com", "iqiyi.com",
            "douyin.com", "weibo.com",
        ]) {
            return (SemanticCategory::Entertainment, 80);
        }

        // 默认浏览
        (SemanticCategory::Browsing, 60)
    }

    /// 辅助：检查文本是否包含任一关键词
    fn has_any_keyword(text: &str, keywords: &[&str]) -> bool {
        keywords.iter().any(|kw| text.contains(kw))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn ide应分类为coding() {
        let result = ClassifierService::classify("Code", "main.rs", None);
        assert_eq!(result.category, Category::Coding);
        assert_eq!(result.semantic_category, SemanticCategory::Coding);
    }

    #[test]
    fn 浏览器应分类为browser() {
        let result = ClassifierService::classify(
            "Chrome",
            "Google",
            Some("https://www.google.com"),
        );
        assert_eq!(result.category, Category::Browser);
    }

    #[test]
    fn github页面应语义分类为coding() {
        let result = ClassifierService::classify(
            "Chrome",
            "rust-lang/rust - GitHub",
            Some("https://github.com/rust-lang/rust"),
        );
        assert_eq!(result.category, Category::Browser);
        assert_eq!(result.semantic_category, SemanticCategory::Coding);
        assert!(result.confidence >= 80);
    }

    #[test]
    fn bilibili应语义分类为entertainment() {
        let result = ClassifierService::classify(
            "Chrome",
            "哔哩哔哩",
            Some("https://www.bilibili.com/video/123"),
        );
        assert_eq!(
            result.semantic_category,
            SemanticCategory::Entertainment,
        );
    }

    #[test]
    fn 微信应分类为communication() {
        let result = ClassifierService::classify("WeChat", "聊天", None);
        assert_eq!(result.category, Category::Communication);
        assert_eq!(
            result.semantic_category,
            SemanticCategory::Communication,
        );
    }

    #[test]
    fn 终端应分类为terminal() {
        let result = ClassifierService::classify(
            "WindowsTerminal",
            "pwsh",
            None,
        );
        assert_eq!(result.category, Category::Terminal);
        assert_eq!(result.semantic_category, SemanticCategory::Coding);
    }

    #[test]
    fn 未知应用应分类为other() {
        let result = ClassifierService::classify("SomeRandomApp", "", None);
        assert_eq!(result.category, Category::Other);
    }

    #[test]
    fn 分类不区分大小写() {
        let result = ClassifierService::classify("CHROME", "Test", None);
        assert_eq!(result.category, Category::Browser);
    }
}

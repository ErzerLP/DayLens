//! # 隐私领域服务
//!
//! 纯同步的业务规则：根据隐私规则配置判断一条活动数据应如何处理。
//! 隐私过滤在数据离开本机前完成，确保敏感信息不会外泄。

use crate::domain::activity::value_objects::PrivacyAction;
use crate::domain::config::entity::PrivacyRule;

/// 隐私领域服务
///
/// 无状态服务，所有方法为关联函数。
/// 接收隐私规则列表和活动数据，返回判定动作。
pub struct PrivacyService;

impl PrivacyService {
    /// 检查一条活动是否匹配隐私规则
    ///
    /// # 参数
    /// - `rules` — 用户配置的隐私规则列表
    /// - `app_name` — 应用名
    /// - `window_title` — 窗口标题
    /// - `browser_url` — 浏览器 URL（可选）
    ///
    /// # 返回
    /// 匹配到的第一条规则的动作，未匹配则返回 `Allow`
    pub fn check_privacy(
        rules: &[PrivacyRule],
        app_name: &str,
        window_title: &str,
        browser_url: Option<&str>,
    ) -> PrivacyAction {
        for rule in rules {
            if Self::matches_rule(rule, app_name, window_title, browser_url) {
                return rule.action.clone();
            }
        }
        PrivacyAction::Allow
    }

    /// 对活动数据执行脱敏处理
    ///
    /// 将窗口标题、浏览器 URL、OCR 文本替换为 `[内容已脱敏]`。
    pub fn anonymize_title(title: &str) -> String {
        if title.is_empty() {
            return title.to_string();
        }
        "[内容已脱敏]".to_string()
    }

    /// 判断单条规则是否匹配
    fn matches_rule(
        rule: &PrivacyRule,
        app_name: &str,
        window_title: &str,
        browser_url: Option<&str>,
    ) -> bool {
        let app_lower = app_name.to_lowercase();
        let title_lower = window_title.to_lowercase();

        // 应用名匹配
        if let Some(ref app_pattern) = rule.app_name_pattern {
            if app_lower.contains(&app_pattern.to_lowercase()) {
                return true;
            }
        }

        // 关键词匹配（标题中包含敏感关键词）
        if let Some(ref keywords) = rule.title_keywords {
            for keyword in keywords {
                if title_lower.contains(&keyword.to_lowercase()) {
                    return true;
                }
            }
        }

        // 域名匹配
        if let Some(ref domain_patterns) = rule.domain_patterns {
            if let Some(url) = browser_url {
                let url_lower = url.to_lowercase();
                for pattern in domain_patterns {
                    if url_lower.contains(&pattern.to_lowercase()) {
                        return true;
                    }
                }
            }
        }

        false
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_rule(
        app: Option<&str>,
        keywords: Option<Vec<&str>>,
        domains: Option<Vec<&str>>,
        action: PrivacyAction,
    ) -> PrivacyRule {
        PrivacyRule {
            app_name_pattern: app.map(String::from),
            title_keywords: keywords.map(|v| {
                v.into_iter().map(String::from).collect()
            }),
            domain_patterns: domains.map(|v| {
                v.into_iter().map(String::from).collect()
            }),
            action,
        }
    }

    #[test]
    fn 无规则时应返回allow() {
        let result = PrivacyService::check_privacy(
            &[],
            "Code",
            "main.rs",
            None,
        );
        assert_eq!(result, PrivacyAction::Allow);
    }

    #[test]
    fn 应用名匹配时应返回对应动作() {
        let rules = vec![
            make_rule(Some("wechat"), None, None, PrivacyAction::Skip),
        ];
        let result = PrivacyService::check_privacy(
            &rules,
            "WeChat",
            "聊天窗口",
            None,
        );
        assert_eq!(result, PrivacyAction::Skip);
    }

    #[test]
    fn 应用名不匹配时应返回allow() {
        let rules = vec![
            make_rule(Some("wechat"), None, None, PrivacyAction::Skip),
        ];
        let result = PrivacyService::check_privacy(
            &rules,
            "Code",
            "main.rs",
            None,
        );
        assert_eq!(result, PrivacyAction::Allow);
    }

    #[test]
    fn 标题关键词匹配时应返回anonymize() {
        let rules = vec![
            make_rule(
                None,
                Some(vec!["密码", "银行"]),
                None,
                PrivacyAction::Anonymize,
            ),
        ];
        let result = PrivacyService::check_privacy(
            &rules,
            "Chrome",
            "修改密码 - 网页",
            None,
        );
        assert_eq!(result, PrivacyAction::Anonymize);
    }

    #[test]
    fn 域名匹配时应返回skip() {
        let rules = vec![
            make_rule(
                None,
                None,
                Some(vec!["bank.com", "alipay.com"]),
                PrivacyAction::Skip,
            ),
        ];
        let result = PrivacyService::check_privacy(
            &rules,
            "Chrome",
            "个人中心",
            Some("https://www.bank.com/account"),
        );
        assert_eq!(result, PrivacyAction::Skip);
    }

    #[test]
    fn 脱敏应替换为固定文本() {
        let result = PrivacyService::anonymize_title("敏感标题");
        assert_eq!(result, "[内容已脱敏]");
    }

    #[test]
    fn 空标题脱敏应保持空() {
        let result = PrivacyService::anonymize_title("");
        assert_eq!(result, "");
    }

    #[test]
    fn 多规则应匹配第一条() {
        let rules = vec![
            make_rule(Some("wechat"), None, None, PrivacyAction::Skip),
            make_rule(
                None,
                Some(vec!["聊天"]),
                None,
                PrivacyAction::Anonymize,
            ),
        ];
        // 两条都能匹配"wechat"，但应返回第一条的 Skip
        let result = PrivacyService::check_privacy(
            &rules,
            "WeChat",
            "聊天窗口",
            None,
        );
        assert_eq!(result, PrivacyAction::Skip);
    }
}

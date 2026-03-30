//! # 同步策略
//!
//! 定义离线缓冲重试的策略：指数退避、最大重试次数、去重键。

use std::time::Duration;

// ===== 常量 =====

/// 基础退避间隔（秒）
const BASE_BACKOFF_SECS: u64 = 1;

/// 最大退避间隔（秒）
const MAX_BACKOFF_SECS: u64 = 300;

/// 最大重试次数
const MAX_RETRY_COUNT: u32 = 20;

/// 重试策略
///
/// 管理缓冲队列中任务的重试逻辑：指数退避 + 最大次数限制。
pub struct RetryPolicy;

impl RetryPolicy {
    /// 计算第 N 次重试后的退避等待时间
    ///
    /// 使用指数退避：`min(base * 2^attempt, max)`
    ///
    /// # 示例
    /// - 第 0 次 → 1s
    /// - 第 1 次 → 2s
    /// - 第 2 次 → 4s
    /// - 第 8 次 → 256s
    /// - 第 9 次 → 300s（截断到最大值）
    pub fn calculate_backoff(attempt: u32) -> Duration {
        let shift = attempt.min(63);
        let multiplier = 1u64.checked_shl(shift).unwrap_or(u64::MAX);
        let secs = BASE_BACKOFF_SECS.saturating_mul(multiplier);
        Duration::from_secs(secs.min(MAX_BACKOFF_SECS))
    }

    /// 判断是否已超过最大重试次数
    pub fn is_exhausted(retry_count: u32) -> bool {
        retry_count >= MAX_RETRY_COUNT
    }

    /// 获取最大重试次数
    pub fn max_retries() -> u32 {
        MAX_RETRY_COUNT
    }
}

/// 去重键 — 由 client_id + client_ts 组成，确保幂等上报
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct DeduplicationKey {
    /// 客户端设备 ID
    pub client_id: String,
    /// 客户端采集时间戳
    pub client_ts: i64,
}

impl DeduplicationKey {
    /// 创建去重键
    pub fn new(client_id: String, client_ts: i64) -> Self {
        Self {
            client_id,
            client_ts,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 第0次退避应为1秒() {
        let backoff = RetryPolicy::calculate_backoff(0);
        assert_eq!(backoff, Duration::from_secs(1));
    }

    #[test]
    fn 第1次退避应为2秒() {
        let backoff = RetryPolicy::calculate_backoff(1);
        assert_eq!(backoff, Duration::from_secs(2));
    }

    #[test]
    fn 第3次退避应为8秒() {
        let backoff = RetryPolicy::calculate_backoff(3);
        assert_eq!(backoff, Duration::from_secs(8));
    }

    #[test]
    fn 超过上限应截断为300秒() {
        let backoff = RetryPolicy::calculate_backoff(10);
        assert_eq!(backoff, Duration::from_secs(300));

        let backoff_large = RetryPolicy::calculate_backoff(30);
        assert_eq!(backoff_large, Duration::from_secs(300));
    }

    #[test]
    fn 重试次数未耗尽() {
        assert!(!RetryPolicy::is_exhausted(0));
        assert!(!RetryPolicy::is_exhausted(19));
    }

    #[test]
    fn 重试次数已耗尽() {
        assert!(RetryPolicy::is_exhausted(20));
        assert!(RetryPolicy::is_exhausted(100));
    }

    #[test]
    fn 去重键相等性() {
        let key1 = DeduplicationKey::new("device-1".to_string(), 12345);
        let key2 = DeduplicationKey::new("device-1".to_string(), 12345);
        let key3 = DeduplicationKey::new("device-2".to_string(), 12345);

        assert_eq!(key1, key2);
        assert_ne!(key1, key3);
    }
}

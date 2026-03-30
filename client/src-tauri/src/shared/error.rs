//! # 共享错误模块
//!
//! 定义全局 `AppError` 枚举和 `Result` 类型别名，
//! 所有 DDD 层统一使用，避免各层自定义错误类型造成混乱。

use serde::Serialize;

/// 应用级错误 — 所有层统一使用
///
/// 每个变体对应一类可区分的故障来源，
/// 便于上层按类型做降级或重试决策。
#[derive(thiserror::Error, Debug)]
pub enum AppError {
    /// 网络请求失败（HTTP 超时、连接拒绝）
    #[error("网络错误: {0}")]
    Network(String),

    /// 服务端返回非 2xx 状态码
    #[error("服务端错误 [{code}]: {message}")]
    Server {
        /// HTTP 状态码
        code: u16,
        /// 服务端返回的错误信息
        message: String,
    },

    /// 离线缓冲操作失败
    #[error("同步错误: {0}")]
    Sync(String),

    /// SQLite 操作失败
    #[error("数据库错误: {0}")]
    Database(String),

    /// 配置文件读写失败
    #[error("配置错误: {0}")]
    Config(String),

    /// 文件 I/O 失败
    #[error("IO 错误: {0}")]
    Io(#[from] std::io::Error),

    /// 截屏/OCR/系统 API 调用失败
    #[error("系统调用错误: {0}")]
    Platform(String),

    /// 通用错误
    #[error("{0}")]
    General(String),
}

/// 统一 Result 类型别名
pub type Result<T> = std::result::Result<T, AppError>;

/// 使 `AppError` 可作为 Tauri 命令的错误返回值
///
/// Tauri 要求 invoke 返回的错误实现 `Serialize`，
/// 这里将错误序列化为一个包含 `error` 字段的 JSON 对象。
impl Serialize for AppError {
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 网络错误应包含描述信息() {
        let err = AppError::Network("连接超时".to_string());
        assert!(err.to_string().contains("连接超时"));
    }

    #[test]
    fn 服务端错误应包含状态码和消息() {
        let err = AppError::Server {
            code: 500,
            message: "内部错误".to_string(),
        };
        let msg = err.to_string();
        assert!(msg.contains("500"));
        assert!(msg.contains("内部错误"));
    }

    #[test]
    fn io错误可自动转换为应用错误() {
        let io_err = std::io::Error::new(std::io::ErrorKind::NotFound, "文件不存在");
        let app_err: AppError = io_err.into();
        assert!(matches!(app_err, AppError::Io(_)));
    }

    #[test]
    fn 错误可序列化为json字符串() {
        let err = AppError::General("测试错误".to_string());
        let json = serde_json::to_string(&err).expect("序列化失败");
        assert!(json.contains("测试错误"));
    }
}

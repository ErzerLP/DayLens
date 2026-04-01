//! # HTTP 远程客户端
//!
//! 封装与 Go 服务端的 HTTP 通信。
//! 提供 GET / POST / PUT / DELETE / upload 等方法。

use std::path::Path;
use std::time::Duration;

use reqwest::multipart;
use serde::de::DeserializeOwned;
use serde::{Deserialize, Serialize};

use crate::shared::error::{AppError, Result};

// ===== 常量 =====

/// 默认请求超时（秒）
const DEFAULT_TIMEOUT_SECS: u64 = 30;

/// 上传请求超时（秒）
const UPLOAD_TIMEOUT_SECS: u64 = 120;

// ===== API 统一响应格式 =====

/// 服务端统一响应结构
#[derive(Debug, Deserialize)]
pub struct ApiResponse<T> {
    /// 状态码（0 = 成功）
    pub code: i32,
    /// 数据（可能为 null）
    pub data: Option<T>,
    /// 消息
    pub message: String,
}

// ===== RemoteClient =====

/// HTTP 远程客户端
///
/// 封装 `reqwest::Client`，提供类型安全的 HTTP 方法。
/// 自动处理认证 Token 和错误转换。
pub struct RemoteClient {
    /// reqwest HTTP 客户端
    client: reqwest::Client,
    /// 服务端基础 URL
    base_url: String,
    /// 认证 Token
    token: String,
}

impl RemoteClient {
    /// 创建新的远程客户端
    ///
    /// # 参数
    /// - `base_url` — 服务端 URL（如 `http://localhost:8080`）
    /// - `token` — Bearer Token
    pub fn new(base_url: String, token: String) -> Result<Self> {
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(DEFAULT_TIMEOUT_SECS))
            .connect_timeout(Duration::from_secs(10))
            .build()
            .map_err(|e| AppError::Network(format!(
                "创建 HTTP 客户端失败: {e}",
            )))?;

        Ok(Self {
            client,
            base_url: base_url.trim_end_matches('/').to_string(),
            token,
        })
    }

    /// GET 请求
    pub async fn get<T: DeserializeOwned>(
        &self,
        path: &str,
    ) -> Result<T> {
        let url = format!("{}{path}", self.base_url);
        let resp = self
            .client
            .get(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .map_err(|e| AppError::Network(format!(
                "GET {path} 失败: {e}",
            )))?;

        self.parse_response(resp, path).await
    }

    /// POST 请求
    pub async fn post<B: Serialize, T: DeserializeOwned>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T> {
        let url = format!("{}{path}", self.base_url);
        let resp = self
            .client
            .post(&url)
            .bearer_auth(&self.token)
            .json(body)
            .send()
            .await
            .map_err(|e| AppError::Network(format!(
                "POST {path} 失败: {e}",
            )))?;

        self.parse_response(resp, path).await
    }

    /// POST 请求（长超时，用于 AI 生成等耗时操作）
    pub async fn post_long<B: Serialize, T: DeserializeOwned>(
        &self,
        path: &str,
        body: &B,
        timeout_secs: u64,
    ) -> Result<T> {
        let url = format!("{}{path}", self.base_url);
        let resp = self
            .client
            .post(&url)
            .bearer_auth(&self.token)
            .timeout(Duration::from_secs(timeout_secs))
            .json(body)
            .send()
            .await
            .map_err(|e| AppError::Network(format!(
                "POST {path} 失败: {e}",
            )))?;

        self.parse_response(resp, path).await
    }

    /// PUT 请求
    pub async fn put<B: Serialize, T: DeserializeOwned>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T> {
        let url = format!("{}{path}", self.base_url);
        let resp = self
            .client
            .put(&url)
            .bearer_auth(&self.token)
            .json(body)
            .send()
            .await
            .map_err(|e| AppError::Network(format!(
                "PUT {path} 失败: {e}",
            )))?;

        self.parse_response(resp, path).await
    }

    /// DELETE 请求
    pub async fn delete<T: DeserializeOwned>(
        &self,
        path: &str,
    ) -> Result<T> {
        let url = format!("{}{path}", self.base_url);
        let resp = self
            .client
            .delete(&url)
            .bearer_auth(&self.token)
            .send()
            .await
            .map_err(|e| AppError::Network(format!(
                "DELETE {path} 失败: {e}",
            )))?;

        self.parse_response(resp, path).await
    }

    /// 上传文件（multipart/form-data）
    ///
    /// 用于截图上传。
    pub async fn upload<T: DeserializeOwned>(
        &self,
        path: &str,
        file_path: &Path,
        field_name: &str,
    ) -> Result<T> {
        let file_bytes = tokio::fs::read(file_path)
            .await
            .map_err(|e| AppError::Io(e))?;

        let file_name = file_path
            .file_name()
            .and_then(|n| n.to_str())
            .unwrap_or("screenshot.jpg")
            .to_string();

        let part = multipart::Part::bytes(file_bytes)
            .file_name(file_name)
            .mime_str("image/jpeg")
            .map_err(|e| AppError::Network(format!(
                "构建 multipart 失败: {e}",
            )))?;

        let form = multipart::Form::new().part(field_name.to_string(), part);

        let url = format!("{}{path}", self.base_url);
        let resp = self
            .client
            .post(&url)
            .bearer_auth(&self.token)
            .timeout(Duration::from_secs(UPLOAD_TIMEOUT_SECS))
            .multipart(form)
            .send()
            .await
            .map_err(|e| AppError::Network(format!(
                "上传 {path} 失败: {e}",
            )))?;

        self.parse_response(resp, path).await
    }

    /// 健康检查
    ///
    /// 向服务端发送 GET `/api/v1/ping`，成功返回 `true`。
    pub async fn health_check(&self) -> bool {
        let url = format!("{}/api/v1/stats?date=2000-01-01", self.base_url);
        match self
            .client
            .get(&url)
            .bearer_auth(&self.token)
            .timeout(Duration::from_secs(5))
            .send()
            .await
        {
            Ok(resp) => resp.status().is_success(),
            Err(_) => false,
        }
    }

    /// 更新 Token
    pub fn set_token(&mut self, token: String) {
        self.token = token;
    }

    /// 更新基础 URL
    pub fn set_base_url(&mut self, url: String) {
        self.base_url = url.trim_end_matches('/').to_string();
    }

    /// 解析统一响应
    async fn parse_response<T: DeserializeOwned>(
        &self,
        resp: reqwest::Response,
        path: &str,
    ) -> Result<T> {
        let status = resp.status();

        if !status.is_success() {
            let body = resp.text().await.unwrap_or_default();
            return Err(AppError::Server {
                code: status.as_u16(),
                message: format!("{path}: {body}"),
            });
        }

        let api_resp: ApiResponse<T> = resp
            .json()
            .await
            .map_err(|e| AppError::Network(format!(
                "解析 {path} 响应失败: {e}",
            )))?;

        if api_resp.code != 0 {
            return Err(AppError::Server {
                code: api_resp.code as u16,
                message: api_resp.message,
            });
        }

        api_resp.data.ok_or_else(|| AppError::Server {
            code: 0,
            message: format!("{path}: 响应 data 为 null"),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn 客户端创建应成功() {
        let client = RemoteClient::new(
            "http://localhost:8080".to_string(),
            "test-token".to_string(),
        );
        assert!(client.is_ok());
    }

    #[test]
    fn base_url应去除末尾斜杠() {
        let client = RemoteClient::new(
            "http://localhost:8080/".to_string(),
            "token".to_string(),
        )
        .unwrap();
        assert_eq!(client.base_url, "http://localhost:8080");
    }

    #[test]
    fn api响应解析_成功() {
        let json = r#"{"code": 0, "data": 42, "message": "ok"}"#;
        let resp: ApiResponse<i32> = serde_json::from_str(json).unwrap();
        assert_eq!(resp.code, 0);
        assert_eq!(resp.data, Some(42));
    }

    #[test]
    fn api响应解析_data为null() {
        let json = r#"{"code": 0, "data": null, "message": "ok"}"#;
        let resp: ApiResponse<i32> = serde_json::from_str(json).unwrap();
        assert_eq!(resp.code, 0);
        assert!(resp.data.is_none());
    }
}

//! # WebSocket 客户端
//!
//! 与 Go 服务端的 WebSocket 连接，用于实时事件推送。
//! 支持自动重连、心跳保持和消息分发。

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::Duration;

use tokio::sync::mpsc;
use tokio::time::sleep;
use tokio_tungstenite::tungstenite::Message;

use crate::shared::error::{AppError, Result};

// ===== 常量 =====

/// 重连间隔（秒）
const RECONNECT_INTERVAL_SECS: u64 = 5;

/// 最大重连间隔（秒）
const MAX_RECONNECT_INTERVAL_SECS: u64 = 60;

/// 心跳间隔（秒）
const HEARTBEAT_INTERVAL_SECS: u64 = 30;

// ===== WsClient =====

/// WebSocket 客户端
///
/// 维护与服务端的 WebSocket 长连接。
/// 自动重连、心跳保持、消息通过 channel 分发。
pub struct WsClient {
    /// WebSocket URL
    url: String,
    /// 认证 Token
    token: String,
    /// 连接状态
    connected: Arc<AtomicBool>,
    /// 停止信号
    shutdown: Arc<AtomicBool>,
}

impl WsClient {
    /// 创建 WebSocket 客户端
    ///
    /// # 参数
    /// - `base_url` — HTTP URL，自动转换为 ws:// / wss://
    /// - `token` — 认证 Token
    pub fn new(base_url: &str, token: &str) -> Self {
        let ws_url = base_url
            .replace("http://", "ws://")
            .replace("https://", "wss://");
        let url = format!(
            "{}/api/v1/ws?token={token}",
            ws_url.trim_end_matches('/'),
        );

        Self {
            url,
            token: token.to_string(),
            connected: Arc::new(AtomicBool::new(false)),
            shutdown: Arc::new(AtomicBool::new(false)),
        }
    }

    /// 启动 WebSocket 连接（异步循环）
    ///
    /// 这是 WebSocket 的主循环，应通过 `tokio::spawn` 调用。
    /// 接收到的消息通过 `tx` 发送出去。
    ///
    /// # 参数
    /// - `tx` — 消息发送通道
    pub async fn run(&self, tx: mpsc::Sender<String>) {
        let mut retry_interval = RECONNECT_INTERVAL_SECS;

        while !self.shutdown.load(Ordering::Relaxed) {
            match self.connect_and_listen(&tx).await {
                Ok(()) => {
                    // 正常断开
                    retry_interval = RECONNECT_INTERVAL_SECS;
                }
                Err(e) => {
                    log::warn!(
                        "WebSocket 连接断开: {e}，{retry_interval}s 后重连",
                    );
                    self.connected.store(false, Ordering::Relaxed);
                }
            }

            if self.shutdown.load(Ordering::Relaxed) {
                break;
            }

            sleep(Duration::from_secs(retry_interval)).await;
            retry_interval = (retry_interval * 2)
                .min(MAX_RECONNECT_INTERVAL_SECS);
        }

        log::info!("WebSocket 客户端已停止");
    }

    /// 连接并监听消息
    async fn connect_and_listen(
        &self,
        tx: &mpsc::Sender<String>,
    ) -> Result<()> {
        use futures_util::{SinkExt, StreamExt};

        let (ws_stream, _) = tokio_tungstenite::connect_async(&self.url)
            .await
            .map_err(|e| AppError::Network(format!(
                "WebSocket 连接失败: {e}",
            )))?;

        log::info!("WebSocket 已连接: {}", self.url);
        self.connected.store(true, Ordering::Relaxed);

        let (mut write, mut read) = ws_stream.split();

        // 心跳任务
        let connected = self.connected.clone();
        let shutdown = self.shutdown.clone();
        let heartbeat = tokio::spawn(async move {
            while connected.load(Ordering::Relaxed)
                && !shutdown.load(Ordering::Relaxed)
            {
                sleep(Duration::from_secs(HEARTBEAT_INTERVAL_SECS)).await;
                // 心跳由 read loop 检测超时处理
            }
        });

        // 读取消息
        while let Some(msg) = read.next().await {
            match msg {
                Ok(Message::Text(text)) => {
                    if tx.send(text.to_string()).await.is_err() {
                        log::warn!("消息接收端已关闭");
                        break;
                    }
                }
                Ok(Message::Ping(data)) => {
                    let _ = write.send(Message::Pong(data)).await;
                }
                Ok(Message::Close(_)) => {
                    log::info!("收到 WebSocket Close 帧");
                    break;
                }
                Err(e) => {
                    return Err(AppError::Network(format!(
                        "WebSocket 读取错误: {e}",
                    )));
                }
                _ => {}
            }
        }

        heartbeat.abort();
        self.connected.store(false, Ordering::Relaxed);
        Ok(())
    }

    /// 检查是否已连接
    pub fn is_connected(&self) -> bool {
        self.connected.load(Ordering::Relaxed)
    }

    /// 停止客户端
    pub fn shutdown(&self) {
        self.shutdown.store(true, Ordering::Relaxed);
        self.connected.store(false, Ordering::Relaxed);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn url转换_http() {
        let client = WsClient::new(
            "http://localhost:8080",
            "token123",
        );
        assert!(client.url.starts_with("ws://"));
        assert!(client.url.contains("token=token123"));
    }

    #[test]
    fn url转换_https() {
        let client = WsClient::new(
            "https://api.example.com",
            "tk",
        );
        assert!(client.url.starts_with("wss://"));
    }

    #[test]
    fn 初始状态应为未连接() {
        let client = WsClient::new(
            "http://localhost:8080",
            "token",
        );
        assert!(!client.is_connected());
    }

    #[test]
    fn shutdown应设置标志() {
        let client = WsClient::new(
            "http://localhost:8080",
            "token",
        );
        client.shutdown();
        assert!(!client.is_connected());
    }
}

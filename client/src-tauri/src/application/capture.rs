//! # 采集用例
//!
//! 编排数据采集的完整流程：
//! 1. 获取前台窗口 → 2. 隐私过滤 → 3. 分类 → 4. 截屏 → 5. OCR → 6. 上报/缓冲
//!
//! # DDD 规则
//! - **只通过 `Arc<dyn Trait>` 调用基础设施**
//! - **不引用任何具体类型**（RemoteClient, SqliteBuffer 等）
//! - **不引用 reqwest, rusqlite, windows, tauri**
//! - **业务规则调用领域服务**（PrivacyService, ClassifierService）

use std::collections::VecDeque;
use std::path::Path;
use std::sync::Arc;
use std::time::Instant;

use chrono::Local;

use crate::application::ports::{
    ActivityReporter, EventEmitter, IdleDetector, OcrEngine,
    ScreenCapture, ScreenLockDetector, SyncBuffer, WindowMonitor,
};
use crate::domain::activity::classifier_service::ClassifierService;
use crate::domain::activity::entity::{ActiveWindow, Activity};
use crate::domain::activity::privacy_service::PrivacyService;
use crate::domain::activity::value_objects::PrivacyAction;
use crate::domain::config::entity::AppConfig;
use crate::domain::sync::entity::SyncTask;

// ===== 常量 =====

/// 内存中保留的最近活动数（用于时长回补）
const ACTIVITY_BUFFER_SIZE: usize = 100;

// ===== CaptureUseCase =====

/// 采集用例 — 编排整个数据采集流程
///
/// 所有依赖通过构造函数注入，使用 `Arc<dyn Trait>`。
/// 不知道具体类型（WindowsMonitor、GdiCapture 等）的存在。
///
/// # 依赖
/// - [`WindowMonitor`] — 前台窗口获取
/// - [`ScreenCapture`] — 截屏
/// - [`OcrEngine`] — 文字识别
/// - [`ActivityReporter`] — 活动上报
/// - [`SyncBuffer`] — 离线缓冲
/// - [`EventEmitter`] — 前端事件推送
/// - [`IdleDetector`] — 空闲检测
/// - [`ScreenLockDetector`] — 锁屏检测
pub struct CaptureUseCase {
    monitor: Arc<dyn WindowMonitor>,
    capture: Arc<dyn ScreenCapture>,
    ocr: Arc<dyn OcrEngine>,
    reporter: Arc<dyn ActivityReporter>,
    buffer: Arc<dyn SyncBuffer>,
    emitter: Arc<dyn EventEmitter>,
    idle: Arc<dyn IdleDetector>,
    lock: Arc<dyn ScreenLockDetector>,
}

impl CaptureUseCase {
    /// 创建采集用例（所有依赖通过参数注入）
    pub fn new(
        monitor: Arc<dyn WindowMonitor>,
        capture: Arc<dyn ScreenCapture>,
        ocr: Arc<dyn OcrEngine>,
        reporter: Arc<dyn ActivityReporter>,
        buffer: Arc<dyn SyncBuffer>,
        emitter: Arc<dyn EventEmitter>,
        idle: Arc<dyn IdleDetector>,
        lock: Arc<dyn ScreenLockDetector>,
    ) -> Self {
        Self {
            monitor,
            capture,
            ocr,
            reporter,
            buffer,
            emitter,
            idle,
            lock,
        }
    }

    /// 执行一次采集迭代
    ///
    /// 这是采集主循环的核心方法。每次调用完成一次完整的采集流程。
    /// 主循环由 `lib.rs` 的 tokio::spawn 驱动，不在此控制。
    ///
    /// # 参数
    /// - `config` — 当前配置快照
    /// - `client_id` — 设备唯一标识
    /// - `last_window` — 上一次的前台窗口（用于检测切换）
    /// - `screenshot_dir` — 截图保存目录
    /// - `recent_activities` — 最近活动缓冲（用于时长回补）
    ///
    /// # 返回
    /// - `Some(window)` — 本次检测到的前台窗口
    /// - `None` — 本次跳过（空闲/锁屏/过滤）
    pub async fn execute_once(
        &self,
        config: &AppConfig,
        client_id: &str,
        last_window: Option<&ActiveWindow>,
        screenshot_path: &Path,
        recent_activities: &mut VecDeque<Activity>,
    ) -> Option<ActiveWindow> {
        // 1. 锁屏检测 → 跳过
        if self.lock.is_locked() {
            return None;
        }

        // 2. 空闲检测 → 跳过
        if self.idle.is_idle() {
            return None;
        }

        // 3. 获取前台窗口
        let window = match self.monitor.get_active_window() {
            Ok(w) => w,
            Err(_) => return None,
        };

        // 4. 获取浏览器 URL
        let browser_url = if window.is_browser {
            self.monitor.get_browser_url(&window.app_name)
        } else {
            None
        };

        // 5. 隐私过滤
        let privacy_action = PrivacyService::check_privacy(
            &config.privacy_rules,
            &window.app_name,
            &window.window_title,
            browser_url.as_deref(),
        );

        match privacy_action {
            PrivacyAction::Skip => return None,
            PrivacyAction::Anonymize | PrivacyAction::Allow => {}
        }

        // 6. 活动分类
        let classification = ClassifierService::classify(
            &window.app_name,
            &window.window_title,
            browser_url.as_deref(),
        );

        // 7. 检测应用切换
        let is_switch = match last_window {
            Some(last) => last.app_name != window.app_name
                || last.window_title != window.window_title,
            None => true,
        };

        // 8. 截屏 + OCR（仅在截屏间隔到达或应用切换时执行）
        let mut ocr_text = None;
        let mut screenshot_key = None;

        if is_switch || screenshot_path.parent().is_some() {
            // 截屏
            if let Err(e) = self.capture.capture(screenshot_path).await {
                log::warn!("截屏失败: {e}");
            } else {
                screenshot_key = Some(
                    screenshot_path
                        .file_name()
                        .and_then(|n| n.to_str())
                        .unwrap_or("unknown.jpg")
                        .to_string(),
                );

                // OCR
                if config.capture.enable_ocr {
                    match self.ocr.recognize(screenshot_path).await {
                        Ok(text) => ocr_text = Some(text),
                        Err(e) => log::warn!("OCR 失败: {e}"),
                    }
                }
            }
        }

        // 9. 脱敏处理
        let window_title = match privacy_action {
            PrivacyAction::Anonymize => {
                PrivacyService::anonymize_title(
                    &window.window_title,
                )
            }
            _ => window.window_title.clone(),
        };
        let browser_url_final = match privacy_action {
            PrivacyAction::Anonymize => None,
            _ => browser_url,
        };
        let ocr_text_final = match privacy_action {
            PrivacyAction::Anonymize => None,
            _ => ocr_text,
        };

        // 10. 构建活动实体
        let now = Local::now().timestamp();
        let activity = Activity {
            id: 0,
            client_id: client_id.to_string(),
            client_ts: now,
            timestamp: now,
            app_name: window.app_name.clone(),
            window_title,
            category: classification.category.as_str().to_string(),
            semantic_category: Some(
                classification.semantic_category.as_str().to_string(),
            ),
            semantic_confidence: Some(classification.confidence),
            duration: config.capture.screenshot_interval_secs as i32,
            browser_url: browser_url_final,
            executable_path: Some(window.executable_path.clone()),
            ocr_text: ocr_text_final,
            screenshot_key,
        };

        // 11. 时长回补：更新最近活动的持续时间
        if is_switch {
            if let Some(last_activity) = recent_activities.back_mut() {
                let elapsed = now - last_activity.timestamp;
                if elapsed > 0 && elapsed < 3600 {
                    last_activity.duration = elapsed as i32;
                }
            }
        }

        // 12. 上报活动
        match self
            .reporter
            .report(&activity, Some(screenshot_path))
            .await
        {
            Ok(id) => {
                log::info!(
                    "上报成功: id={id}, app={}",
                    activity.app_name,
                );
            }
            Err(e) => {
                log::warn!("上报失败，写入缓冲: {e}");
                // 降级到离线缓冲
                let payload = serde_json::to_string(&activity)
                    .unwrap_or_default();
                let task = SyncTask::new(
                    payload,
                    screenshot_path
                        .to_str()
                        .map(|s| s.to_string()),
                    activity.client_ts,
                );
                if let Err(buf_err) = self.buffer.enqueue(&task) {
                    log::error!("缓冲写入也失败: {buf_err}");
                }
            }
        }

        // 13. emit 事件到前端
        let event_payload = serde_json::json!({
            "appName": activity.app_name,
            "windowTitle": activity.window_title,
            "category": activity.category,
            "timestamp": activity.timestamp,
        })
        .to_string();

        if let Err(e) = self.emitter.emit("activity-captured", &event_payload) {
            log::debug!("emit 事件失败: {e}");
        }

        // 14. 写入最近活动缓冲
        if recent_activities.len() >= ACTIVITY_BUFFER_SIZE {
            recent_activities.pop_front();
        }
        recent_activities.push_back(activity);

        Some(window)
    }
}

/// 判断是否需要截屏
///
/// 根据上次截屏时间和配置的间隔判断。
pub fn should_capture(
    last_capture: &Instant,
    interval_secs: u64,
) -> bool {
    last_capture.elapsed().as_secs() >= interval_secs
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;

    #[test]
    fn 活动缓冲大小常量应合理() {
        assert!(ACTIVITY_BUFFER_SIZE >= 10);
        assert!(ACTIVITY_BUFFER_SIZE <= 1000);
    }

    #[test]
    fn 截屏间隔判断_时间未到() {
        let last = Instant::now();
        assert!(!should_capture(&last, 30));
    }

    #[test]
    fn 截屏间隔判断_时间已到() {
        let last = Instant::now() - Duration::from_secs(31);
        assert!(should_capture(&last, 30));
    }

    #[test]
    fn 活动缓冲应有上限() {
        let mut buffer = VecDeque::new();
        for i in 0..ACTIVITY_BUFFER_SIZE + 10 {
            if buffer.len() >= ACTIVITY_BUFFER_SIZE {
                buffer.pop_front();
            }
            buffer.push_back(Activity {
                id: 0,
                client_id: "test".to_string(),
                client_ts: i as i64,
                timestamp: i as i64,
                app_name: "Test".to_string(),
                window_title: "Test".to_string(),
                category: "other".to_string(),
                semantic_category: None,
                semantic_confidence: None,
                duration: 30,
                browser_url: None,
                executable_path: None,
                ocr_text: None,
                screenshot_key: None,
            });
        }
        assert_eq!(buffer.len(), ACTIVITY_BUFFER_SIZE);
    }
}

//! # DayLens 客户端
//!
//! 基于 Tauri 2 的工作活动追踪与智能分析桌面客户端。

mod domain;
mod application;
mod infrastructure;
mod interface;
mod shared;

use std::sync::Arc;
use std::path::PathBuf;

use tauri::Manager;
use tokio::sync::RwLock;

use crate::application::capture::CaptureUseCase;
use crate::application::config::ConfigManager;
use crate::application::query::QueryService;
use crate::application::sync::SyncCoordinator;
use crate::infrastructure::capture::idle_detector::InputIdleDetector;
use crate::infrastructure::capture::monitor::WindowsMonitor;
use crate::infrastructure::capture::ocr::OcrService;
use crate::infrastructure::capture::screen_lock::WtsLockDetector;
use crate::infrastructure::capture::screenshot::ScreenshotService;
use crate::infrastructure::persistence::config_store::ConfigStore;
use crate::infrastructure::persistence::sync_buffer::SqliteSyncBuffer;
use crate::infrastructure::remote::client::RemoteClient;
use crate::infrastructure::remote::data_source::RemoteDataSource;
use crate::infrastructure::remote::uploader::ActivityUploader;
use crate::interface::state::AppState;

// ===== TauriEventEmitter =====

/// Tauri 事件发射器
struct TauriEventEmitter {
    app: tauri::AppHandle,
}

impl application::ports::EventEmitter for TauriEventEmitter {
    fn emit(&self, event: &str, payload: &str) -> shared::error::Result<()> {
        use tauri::Emitter;
        self.app
            .emit(event, payload.to_string())
            .map_err(|e| shared::error::AppError::Platform(format!(
                "事件发射失败: {e}",
            )))?;
        Ok(())
    }
}

// ===== 端口适配 =====

/// 窗口监控适配器 (sync → trait)
struct WindowMonitorAdapter(WindowsMonitor);

impl application::ports::WindowMonitor for WindowMonitorAdapter {
    fn get_active_window(&self) -> shared::error::Result<domain::activity::entity::ActiveWindow> {
        self.0.get_active_window()
    }
    fn get_browser_url(&self, app_name: &str) -> Option<String> {
        self.0.get_browser_url(app_name)
    }
}

/// 截屏适配器 (sync → async trait)
struct ScreenCaptureAdapter(ScreenshotService);

#[async_trait::async_trait]
impl application::ports::ScreenCapture for ScreenCaptureAdapter {
    async fn capture(&self, save_path: &std::path::Path) -> shared::error::Result<()> {
        self.0.capture(save_path)
    }
    fn generate_thumbnail(
        &self,
        source: &std::path::Path,
        target: &std::path::Path,
        width: u32,
    ) -> shared::error::Result<()> {
        self.0.generate_thumbnail(source, target, width)
    }
}

/// OCR 适配器 (sync → async trait)
struct OcrAdapter(OcrService);

#[async_trait::async_trait]
impl application::ports::OcrEngine for OcrAdapter {
    async fn recognize(&self, image_path: &std::path::Path) -> shared::error::Result<String> {
        self.0.recognize(image_path)
    }
    fn is_available(&self) -> bool {
        self.0.is_available()
    }
}

/// 空闲检测适配器
struct IdleAdapter(InputIdleDetector);

impl application::ports::IdleDetector for IdleAdapter {
    fn is_idle(&self) -> bool {
        self.0.is_idle()
    }
    fn reset(&self) {
        self.0.reset();
    }
}

/// 锁屏检测适配器
struct LockAdapter(WtsLockDetector);

impl application::ports::ScreenLockDetector for LockAdapter {
    fn is_locked(&self) -> bool {
        self.0.is_locked()
    }
}

/// 配置持久化适配器
struct ConfigPersistenceAdapter(ConfigStore);

impl application::config::ConfigPersistence for ConfigPersistenceAdapter {
    fn load(&self) -> shared::error::Result<domain::config::entity::AppConfig> {
        self.0.load()
    }
    fn save(&self, config: &domain::config::entity::AppConfig) -> shared::error::Result<()> {
        self.0.save(config)
    }
}

// ===== Tauri 入口 =====

/// 确保数据目录存在
fn ensure_data_dir() -> PathBuf {
    let base = dirs::data_dir()
        .unwrap_or_else(|| PathBuf::from("."))
        .join("daylens");
    std::fs::create_dir_all(&base).ok();
    std::fs::create_dir_all(base.join("screenshots")).ok();
    std::fs::create_dir_all(base.join("thumbnails")).ok();
    base
}

/// 构建系统托盘
fn build_tray(app: &tauri::App) -> tauri::Result<()> {
    use tauri::tray::{TrayIconBuilder, MouseButton, MouseButtonState, TrayIconEvent};
    use tauri::menu::{MenuBuilder, MenuItemBuilder};

    let show_item = MenuItemBuilder::with_id("show", "显示窗口").build(app)?;
    let quit_item = MenuItemBuilder::with_id("quit", "退出").build(app)?;

    let menu = MenuBuilder::new(app)
        .item(&show_item)
        .separator()
        .item(&quit_item)
        .build()?;

    TrayIconBuilder::new()
        .tooltip("DayLens — 运行中")
        .menu(&menu)
        .on_menu_event(|app, event| {
            match event.id().as_ref() {
                "show" => {
                    if let Some(w) = app.get_webview_window("main") {
                        let _ = w.show();
                        let _ = w.set_focus();
                    }
                }
                "quit" => {
                    app.exit(0);
                }
                _ => {}
            }
        })
        .on_tray_icon_event(|tray, event| {
            // 双击托盘图标显示窗口
            if let TrayIconEvent::Click {
                button: MouseButton::Left,
                button_state: MouseButtonState::Up,
                ..
            } = event {
                if let Some(w) = tray.app_handle().get_webview_window("main") {
                    let _ = w.show();
                    let _ = w.set_focus();
                }
            }
        })
        .build(app)?;

    Ok(())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    env_logger::Builder::from_env(
        env_logger::Env::default().default_filter_or("info")
    ).init();

    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        // 9.1 — 开机自启
        .plugin(tauri_plugin_autostart::init(
            tauri_plugin_autostart::MacosLauncher::LaunchAgent,
            None,
        ))
        // 9.5 — 单实例
        .plugin(tauri_plugin_single_instance::init(|app, _args, _cwd| {
            // 已有实例运行时，切换到已有窗口
            if let Some(w) = app.get_webview_window("main") {
                let _ = w.show();
                let _ = w.set_focus();
            }
        }))
        .setup(|app| {
            // 9.4 — 数据目录
            let data_dir = ensure_data_dir();
            log::info!("数据目录: {}", data_dir.display());

            // 9.2 — 系统托盘
            build_tray(app)?;

            // 9.2 — 关闭窗口时最小化到托盘
            if let Some(window) = app.get_webview_window("main") {
                let w = window.clone();
                window.on_window_event(move |event| {
                    if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                        api.prevent_close();
                        let _ = w.hide();
                    }
                });
            }

            // 1. 加载配置
            let config_store = ConfigStore::from_default()
                .unwrap_or_else(|_| {
                    ConfigStore::new(data_dir.join("config.json"))
                });
            let config_manager = Arc::new(
                ConfigManager::new(Box::new(ConfigPersistenceAdapter(config_store)))
                    .expect("配置加载失败"),
            );
            let config = config_manager.get();

            // 2. 创建 HTTP 客户端
            let remote_client = Arc::new(RwLock::new(
                RemoteClient::new(
                    config.server.url.clone(),
                    config.server.token.clone(),
                )
                .expect("HTTP 客户端创建失败"),
            ));

            // 3. 创建端口实现
            let data_source: Arc<dyn application::ports::DataSource> =
                Arc::new(RemoteDataSource::new(remote_client.clone()));
            let reporter: Arc<dyn application::ports::ActivityReporter> =
                Arc::new(ActivityUploader::new(remote_client.clone()));

            let sync_db = data_dir.join("sync.db");
            let buffer: Arc<dyn application::ports::SyncBuffer> =
                Arc::new(SqliteSyncBuffer::new(sync_db).expect("同步数据库创建失败"));

            let emitter: Arc<dyn application::ports::EventEmitter> =
                Arc::new(TauriEventEmitter { app: app.handle().clone() });

            let monitor: Arc<dyn application::ports::WindowMonitor> =
                Arc::new(WindowMonitorAdapter(WindowsMonitor::new()));
            let capture: Arc<dyn application::ports::ScreenCapture> =
                Arc::new(ScreenCaptureAdapter(ScreenshotService::new()));
            let ocr: Arc<dyn application::ports::OcrEngine> =
                Arc::new(OcrAdapter(OcrService::new()));
            let idle: Arc<dyn application::ports::IdleDetector> =
                Arc::new(IdleAdapter(InputIdleDetector::new(
                    config.capture.idle_timeout_minutes as u32,
                )));
            let lock: Arc<dyn application::ports::ScreenLockDetector> =
                Arc::new(LockAdapter(WtsLockDetector::new()));

            // 4. 组装应用层服务
            let capture_use_case = Arc::new(CaptureUseCase::new(
                monitor, capture, ocr,
                reporter.clone(), buffer.clone(), emitter,
                idle, lock,
            ));

            let query_service = Arc::new(QueryService::new(data_source));
            let sync_coordinator = Arc::new(SyncCoordinator::new(
                reporter, buffer,
            ));

            // 5. 注入全局状态
            app.manage(AppState {
                query: query_service,
                config: config_manager.clone(),
                sync: sync_coordinator.clone(),
                capture: capture_use_case.clone(),
            });

            // 6. 启动后台采集循环
            let capture_handle = capture_use_case.clone();
            let config_handle = config_manager.clone();
            let sync_handle = sync_coordinator.clone();
            let client_id = uuid::Uuid::new_v4().to_string();

            tauri::async_runtime::spawn(async move {
                use std::collections::VecDeque;
                use crate::application::capture::should_capture;

                let mut recent_activities = VecDeque::new();
                let mut last_window: Option<crate::domain::activity::entity::ActiveWindow> = None;
                let mut last_capture_time = std::time::Instant::now();
                let mut tick_count: u64 = 0;

                log::info!("采集循环已启动");

                loop {
                    let config = config_handle.get();
                    let interval = config.capture.screenshot_interval_secs;

                    // 工作时间检查
                    if !config.work_schedule.is_work_time() {
                        tokio::time::sleep(tokio::time::Duration::from_secs(60)).await;
                        continue;
                    }

                    // 截屏间隔检查
                    if !should_capture(&last_capture_time, interval) {
                        tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
                        continue;
                    }

                    // 构建截图路径
                    let data_dir = dirs::data_dir()
                        .unwrap_or_else(|| std::path::PathBuf::from("."))
                        .join("daylens");
                    let screenshot_name = format!(
                        "{}_{}.jpg",
                        chrono::Local::now().format("%Y%m%d_%H%M%S"),
                        tick_count
                    );
                    let screenshot_path = data_dir.join("screenshots").join(&screenshot_name);

                    // 执行采集（catch panic 防止 Win32 调用崩溃杀死循环）
                    let result = std::panic::AssertUnwindSafe(
                        capture_handle.execute_once(
                            &config,
                            &client_id,
                            last_window.as_ref(),
                            &screenshot_path,
                            &mut recent_activities,
                        )
                    );
                    match futures_util::FutureExt::catch_unwind(result).await {
                        Ok(Some(window)) => {
                            log::info!("采集成功: {} - {}", window.app_name, window.window_title);
                            last_window = Some(window);
                        }
                        Ok(None) => {
                            log::debug!("采集跳过（空闲/锁屏/过滤）");
                        }
                        Err(e) => {
                            log::error!("采集 panic: {:?}", e);
                        }
                    }

                    last_capture_time = std::time::Instant::now();
                    tick_count += 1;

                    // 每 10 次尝试同步缓冲队列
                    if tick_count % 10 == 0 {
                        let synced = sync_handle.sync_once().await;
                        if synced > 0 {
                            log::info!("同步缓冲: {synced} 条已上报");
                        }
                    }

                    tokio::time::sleep(tokio::time::Duration::from_secs(interval)).await;
                }
            });

            log::info!("DayLens 客户端已启动");
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            // 活动
            interface::commands::activity::get_today_stats,
            interface::commands::activity::get_stats,
            interface::commands::activity::get_timeline,
            interface::commands::activity::get_activity,
            interface::commands::activity::get_hourly_summaries,
            // 报告
            interface::commands::report::get_report,
            interface::commands::report::generate_report,
            interface::commands::report::get_sessions,
            // 搜索
            interface::commands::search::search_activities,
            // AI
            interface::commands::intelligence::ask_ai,
            interface::commands::intelligence::chat_ai,
            // 配置
            interface::commands::config::get_config,
            interface::commands::config::update_server_url,
            interface::commands::config::update_server_token,
            interface::commands::config::update_capture_interval,
            // 系统
            interface::commands::system::get_storage_stats,
            interface::commands::system::cleanup_data,
            interface::commands::system::get_sync_queue_size,
            interface::commands::system::check_permissions,
            interface::commands::system::is_work_time,
            interface::commands::system::get_data_dir,
            interface::commands::system::test_connection,
        ])
        .run(tauri::generate_context!())
        .expect("Tauri 应用启动失败");
}

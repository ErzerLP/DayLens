//! 应用全局状态，注入到 Tauri `manage()`

use std::sync::Arc;

use crate::application::capture::CaptureUseCase;
use crate::application::config::ConfigManager;
use crate::application::query::QueryService;
use crate::application::sync::SyncCoordinator;

/// 全局应用状态
///
/// 通过 `tauri::State<AppState>` 在命令中访问。
pub struct AppState {
    pub query: Arc<QueryService>,
    pub config: Arc<ConfigManager>,
    pub sync: Arc<SyncCoordinator>,
    pub capture: Arc<CaptureUseCase>,
}

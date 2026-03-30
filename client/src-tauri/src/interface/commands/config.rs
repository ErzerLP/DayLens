//! 配置管理命令

use tauri::State;

use crate::domain::config::entity::AppConfig;
use crate::interface::state::AppState;

#[tauri::command]
pub async fn get_config(
    state: State<'_, AppState>,
) -> Result<AppConfig, String> {
    Ok(state.config.get())
}

#[tauri::command]
pub async fn update_server_url(
    state: State<'_, AppState>,
    url: String,
) -> Result<(), String> {
    state
        .config
        .update(|c| c.server.url = url)
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn update_server_token(
    state: State<'_, AppState>,
    token: String,
) -> Result<(), String> {
    state
        .config
        .update(|c| c.server.token = token)
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn update_capture_interval(
    state: State<'_, AppState>,
    secs: u64,
) -> Result<(), String> {
    state
        .config
        .update(|c| c.capture.screenshot_interval_secs = secs)
        .map_err(|e| e.to_string())
}

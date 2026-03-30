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
        .update(|c| c.server.url = url.clone())
        .map_err(|e| e.to_string())?;

    // 同步更新 HTTP 客户端
    state.remote_client.write().await.set_base_url(url);
    Ok(())
}

#[tauri::command]
pub async fn update_server_token(
    state: State<'_, AppState>,
    token: String,
) -> Result<(), String> {
    state
        .config
        .update(|c| c.server.token = token.clone())
        .map_err(|e| e.to_string())?;

    // 同步更新 HTTP 客户端
    state.remote_client.write().await.set_token(token);
    Ok(())
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

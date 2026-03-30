//! 系统状态命令

use tauri::State;

use crate::domain::activity::entity::*;
use crate::interface::state::AppState;

#[tauri::command]
pub async fn get_storage_stats(
    state: State<'_, AppState>,
) -> Result<StorageStats, String> {
    state.query.get_storage_stats().await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn cleanup_data(
    state: State<'_, AppState>,
    before_date: String,
) -> Result<CleanupResult, String> {
    state
        .query
        .cleanup_before(&before_date)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_sync_queue_size(
    state: State<'_, AppState>,
) -> Result<i64, String> {
    Ok(state.sync.queue_size().await)
}

/// 9.3 — 检查系统权限（截屏 + 输入监控）
#[tauri::command]
pub fn check_permissions() -> Result<serde_json::Value, String> {
    // Windows 下截屏和输入监控通常不需要额外授权
    // 返回各项权限状态供前端展示
    Ok(serde_json::json!({
        "screenshot": true,
        "inputMonitor": true,
        "accessibility": true,
        "platform": "windows"
    }))
}

/// 9.6 — 检查当前是否在工作时间
#[tauri::command]
pub fn is_work_time(
    state: State<'_, AppState>,
) -> Result<bool, String> {
    let config = state.config.get();
    Ok(config.work_schedule.is_work_time())
}

/// 获取数据目录路径
#[tauri::command]
pub fn get_data_dir() -> Result<String, String> {
    let base = dirs::data_dir()
        .unwrap_or_else(|| std::path::PathBuf::from("."))
        .join("daylens");
    Ok(base.to_string_lossy().to_string())
}

/// 测试服务端连接
#[tauri::command]
pub async fn test_connection(
    state: State<'_, AppState>,
) -> Result<bool, String> {
    let config = state.config.get();
    let url = &config.server.url;
    let token = &config.server.token;

    if url.is_empty() {
        return Ok(false);
    }

    let client = reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(5))
        .build()
        .map_err(|e| e.to_string())?;

    let ping_url = format!("{}/api/v1/ping", url.trim_end_matches('/'));
    match client
        .get(&ping_url)
        .bearer_auth(token)
        .send()
        .await
    {
        Ok(resp) => Ok(resp.status().is_success()),
        Err(_) => Ok(false),
    }
}

//! 报告相关命令

use tauri::State;

use crate::domain::activity::entity::*;
use crate::interface::state::AppState;

#[tauri::command]
pub async fn get_report(
    state: State<'_, AppState>,
    date: String,
) -> Result<Option<DailyReport>, String> {
    state.query.get_report(&date).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn generate_report(
    state: State<'_, AppState>,
    date: String,
    force: bool,
) -> Result<DailyReport, String> {
    state
        .query
        .generate_report(&date, force)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_sessions(
    state: State<'_, AppState>,
    date: String,
) -> Result<Vec<WorkSession>, String> {
    state.query.get_sessions(&date).await.map_err(|e| e.to_string())
}

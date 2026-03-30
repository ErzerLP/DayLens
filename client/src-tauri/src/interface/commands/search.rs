//! 搜索命令

use tauri::State;

use crate::domain::activity::entity::*;
use crate::interface::state::AppState;

#[tauri::command]
pub async fn search_activities(
    state: State<'_, AppState>,
    query: String,
    limit: i32,
) -> Result<Vec<SearchResultItem>, String> {
    state.query.search(&query, limit).await.map_err(|e| e.to_string())
}

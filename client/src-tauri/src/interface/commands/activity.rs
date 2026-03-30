//! 活动查询命令

use tauri::State;

use crate::domain::activity::entity::*;
use crate::interface::state::AppState;

#[tauri::command]
pub async fn get_today_stats(
    state: State<'_, AppState>,
) -> Result<DailyStats, String> {
    let date = chrono::Local::now().format("%Y-%m-%d").to_string();
    state.query.get_stats(&date).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_stats(
    state: State<'_, AppState>,
    date: String,
) -> Result<DailyStats, String> {
    state.query.get_stats(&date).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_timeline(
    state: State<'_, AppState>,
    date: String,
    limit: i32,
    offset: i32,
    app: Option<String>,
    category: Option<String>,
) -> Result<TimelineResponse, String> {
    state
        .query
        .get_timeline(
            &date,
            limit,
            offset,
            app.as_deref(),
            category.as_deref(),
        )
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_activity(
    state: State<'_, AppState>,
    id: i64,
) -> Result<Activity, String> {
    state.query.get_activity(id).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_hourly_summaries(
    state: State<'_, AppState>,
    date: String,
) -> Result<Vec<HourlySummary>, String> {
    state
        .query
        .get_hourly_summaries(&date)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn get_app_categories(
    state: State<'_, AppState>,
) -> Result<Vec<AppCategoryInfo>, String> {
    let today = chrono::Local::now().format("%Y-%m-%d").to_string();
    let from = (chrono::Local::now() - chrono::Duration::days(90))
        .format("%Y-%m-%d")
        .to_string();
    state
        .query
        .get_app_categories(&from, &today)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn set_category_rule(
    state: State<'_, AppState>,
    app_name: String,
    category: String,
) -> Result<(), String> {
    state
        .query
        .set_category_rule(&app_name, &category)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn reclassify_app(
    state: State<'_, AppState>,
    app_name: String,
    new_category: String,
) -> Result<i64, String> {
    state
        .query
        .reclassify_app(&app_name, &new_category)
        .await
        .map_err(|e| e.to_string())
}

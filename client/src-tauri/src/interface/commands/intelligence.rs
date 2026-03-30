//! AI 智能相关命令

use tauri::State;

use crate::domain::activity::entity::*;
use crate::interface::state::AppState;

#[tauri::command]
pub async fn ask_ai(
    state: State<'_, AppState>,
    question: String,
    context: String,
) -> Result<AiAnswer, String> {
    state.query.ask(&question, &context).await.map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn chat_ai(
    state: State<'_, AppState>,
    messages: Vec<ChatMessage>,
    tools: Vec<String>,
) -> Result<AssistantReply, String> {
    state.query.chat(messages, tools).await.map_err(|e| e.to_string())
}

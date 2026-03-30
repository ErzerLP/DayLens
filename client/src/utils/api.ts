// Tauri invoke 命令名常量

export const CMD = {
  // 活动
  GET_TODAY_STATS: "get_today_stats",
  GET_STATS: "get_stats",
  GET_TIMELINE: "get_timeline",
  GET_ACTIVITY: "get_activity",
  GET_HOURLY_SUMMARIES: "get_hourly_summaries",
  // 报告
  GET_REPORT: "get_report",
  GENERATE_REPORT: "generate_report",
  GET_SESSIONS: "get_sessions",
  // 搜索
  SEARCH_ACTIVITIES: "search_activities",
  // AI
  ASK_AI: "ask_ai",
  CHAT_AI: "chat_ai",
  // 配置
  GET_CONFIG: "get_config",
  UPDATE_SERVER_URL: "update_server_url",
  UPDATE_SERVER_TOKEN: "update_server_token",
  UPDATE_CAPTURE_INTERVAL: "update_capture_interval",
  // 系统
  GET_STORAGE_STATS: "get_storage_stats",
  CLEANUP_DATA: "cleanup_data",
  GET_SYNC_QUEUE_SIZE: "get_sync_queue_size",
  TEST_CONNECTION: "test_connection",
  // 应用分类
  GET_APP_CATEGORIES: "get_app_categories",
  SET_CATEGORY_RULE: "set_category_rule",
  RECLASSIFY_APP: "reclassify_app",
} as const;

// Tauri 事件名
export const EVENT = {
  ACTIVITY_CAPTURED: "activity-captured",
  CAPTURE_STATUS: "capture-status-changed",
  SYNC_STATUS: "sync-status-changed",
  SCREENSHOT_TAKEN: "screenshot-taken",
} as const;

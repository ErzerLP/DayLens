// DayLens 前端类型定义 — 与 Rust 后端 entity 镜像

export interface DailyStats {
  date: string;
  totalActivities: number;
  totalMinutes: number;
  topApps: AppStat[];
  categoryBreakdown: CategoryBreakdown[];
  firstActivity: string | null;
  lastActivity: string | null;
}

export interface AppStat {
  appName: string;
  count: number;
  minutes: number;
}

export interface CategoryBreakdown {
  category: string;
  minutes: number;
  percentage: number;
}

export interface Activity {
  id: number;
  appName: string;
  windowTitle: string;
  startedAt: string;
  endedAt: string;
  durationSecs: number;
  category: string;
  semanticCategory: string;
  ocrText: string | null;
  screenshotPath: string | null;
  thumbnailPath: string | null;
}

export interface TimelineResponse {
  items: Activity[];
  total: number;
  hasMore: boolean;
}

export interface HourlySummary {
  hour: number;
  totalMinutes: number;
  topApp: string;
  topCategory: string;
  activityCount: number;
}

export interface DailyReport {
  date: string;
  content: string;
  generatedAt: string;
}

export interface WeeklyReview {
  fromDate: string;
  toDate: string;
  content: string;
  generatedAt: string;
}

export interface WorkSession {
  id: number;
  startedAt: string;
  endedAt: string;
  durationMinutes: number;
  dominantApp: string;
  dominantCategory: string;
}

export interface SearchResultItem {
  activityId: number;
  appName: string;
  windowTitle: string;
  matchedText: string;
  timestamp: string;
  score: number;
}

export interface AiAnswer {
  answer: string;
  sources: string[];
}

export interface ChatMessage {
  role: "user" | "assistant";
  content: string;
}

export interface AssistantReply {
  content: string;
  toolCalls: string[];
}

export interface StorageStats {
  totalBytes: number;
  screenshotBytes: number;
  databaseBytes: number;
  screenshotCount: number;
}

export interface CleanupResult {
  deletedScreenshots: number;
  freedBytes: number;
}

export interface AppConfig {
  server: ServerConfig;
  capture: CaptureConfig;
  privacy: PrivacyConfig;
}

export interface ServerConfig {
  url: string;
  token: string;
}

export interface CaptureConfig {
  screenshotIntervalSecs: number;
  idleTimeoutMinutes: number;
  enableOcr: boolean;
  enableScreenshot: boolean;
}

export interface PrivacyConfig {
  enablePrivacyFilter: boolean;
  blockedApps: string[];
}

// DayLens 前端类型定义 — 与 Rust 后端 entity 镜像

export interface DailyStats {
  date: string;
  // 服务端实际字段（秒）
  totalDuration?: number;
  screenshotCount?: number;
  activeHours?: number;
  appUsage?: AppUsage[];
  categoryUsage?: CategoryUsage[];
  workTimeDuration?: number;
  // 前端兼容字段（分钟）
  totalActivities?: number;
  totalMinutes?: number;
  topApps?: AppStat[];
  categoryBreakdown?: CategoryBreakdown[];
  firstActivity?: string | null;
  lastActivity?: string | null;
}

export interface AppUsage {
  appName: string;
  duration: number;
  count: number;
}

export interface CategoryUsage {
  category: string;
  duration: number;
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
  // 服务端字段
  timestamp?: number;
  duration?: number;
  // 客户端兼容字段
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
  hasMore?: boolean;
  limit?: number;
  offset?: number;
}

export interface HourlySummary {
  hour: number;
  totalDuration?: number;
  totalMinutes?: number;
  topApp: string;
  topCategory: string;
  activityCount: number;
}

export interface DailyReport {
  date: string;
  content: string;
  generatedAt: string;
  usedAi?: boolean;
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
  activityCount: number;
  screenshotCount: number;
  diskUsageMb: number;
  maxStorageMb: number;
  oldestActivityDate: string;
  retentionDays: number;
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

// 应用分类信息
export interface AppCategoryInfo {
  appName: string;
  category: string;
  isCustomRule: boolean;
  totalDuration: number;
  lastSeen: number;
}

// 所有可用分类
export const CATEGORIES = [
  { key: "coding", label: "编码开发", color: "hsl(210, 90%, 60%)" },
  { key: "browser", label: "网页浏览", color: "hsl(38, 92%, 55%)" },
  { key: "communication", label: "即时通讯", color: "hsl(145, 65%, 48%)" },
  { key: "document", label: "文档编辑", color: "hsl(260, 70%, 60%)" },
  { key: "design", label: "设计工具", color: "hsl(330, 70%, 55%)" },
  { key: "terminal", label: "终端命令", color: "hsl(180, 50%, 50%)" },
  { key: "media", label: "媒体播放", color: "hsl(0, 75%, 55%)" },
  { key: "system", label: "系统工具", color: "hsl(220, 14%, 50%)" },
  { key: "gaming", label: "游戏娱乐", color: "hsl(290, 60%, 55%)" },
  { key: "other", label: "其他", color: "hsl(0, 0%, 50%)" },
] as const;

export function getCategoryInfo(key: string) {
  return CATEGORIES.find((c) => c.key === key) ?? CATEGORIES[CATEGORIES.length - 1];
}

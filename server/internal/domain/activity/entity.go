// Package activity 活动领域实体。
//
// 定义活动采集、统计、应用使用等核心数据结构。
// 所有结构体均为纯数据，可跨层使用。
package activity

// Activity 一条活动记录 — 系统采集的单次窗口活动快照
type Activity struct {
	ID                  int64   `json:"id"`
	ClientID            string  `json:"clientId"`
	ClientTs            int64   `json:"clientTs"`
	Timestamp           int64   `json:"timestamp"`
	AppName             string  `json:"appName"`
	WindowTitle         string  `json:"windowTitle"`
	Category            string  `json:"category"`
	SemanticCategory    *string `json:"semanticCategory"`
	SemanticConfidence  *int    `json:"semanticConfidence"`
	Duration            int     `json:"duration"`
	BrowserURL          *string `json:"browserUrl"`
	ExecutablePath      *string `json:"executablePath"`
	OcrText             *string `json:"ocrText"`
	ScreenshotKey       *string `json:"screenshotKey"`
}

// DailyStats 每日统计
type DailyStats struct {
	Date             string             `json:"date"`
	TotalDuration    int64              `json:"totalDuration"`
	ScreenshotCount  int64              `json:"screenshotCount"`
	ActiveHours      int                `json:"activeHours"`
	AppUsage         []AppUsage         `json:"appUsage"`
	CategoryUsage    []CategoryUsage    `json:"categoryUsage"`
	DomainUsage      []DomainUsage      `json:"domainUsage"`
	WorkTimeDuration int64              `json:"workTimeDuration"`
	TopWindowTitles  []WindowTitleUsage `json:"topWindowTitles"`
}

// AppUsage 应用使用统计
type AppUsage struct {
	AppName  string `json:"appName"`
	Duration int64  `json:"duration"`
	Count    int64  `json:"count"`
}

// CategoryUsage 分类使用统计
type CategoryUsage struct {
	Category string `json:"category"`
	Duration int64  `json:"duration"`
}

// DomainUsage 域名使用统计
type DomainUsage struct {
	Domain   string `json:"domain"`
	Duration int64  `json:"duration"`
}

// WindowTitleUsage 窗口标题使用统计
type WindowTitleUsage struct {
	Title   string `json:"title"`
	Duration int64  `json:"duration"`
	AppName string `json:"appName"`
}

// HourlySummary 小时摘要
type HourlySummary struct {
	Hour                       int      `json:"hour"`
	Summary                    string   `json:"summary"`
	MainApps                   string   `json:"mainApps"`
	ActivityCount              int64    `json:"activityCount"`
	TotalDuration              int64    `json:"totalDuration"`
	RepresentativeScreenshots  []string `json:"representativeScreenshots"`
}

// SearchResultItem 搜索结果条目
type SearchResultItem struct {
	ActivityID    int64  `json:"activityId"`
	Timestamp     int64  `json:"timestamp"`
	AppName       string `json:"appName"`
	Excerpt       string `json:"excerpt"`
	MatchField    string `json:"matchField"`
	RelevanceScore int64  `json:"relevanceScore"`
}

// AiAnswer AI 问答回答
type AiAnswer struct {
	Answer     string        `json:"answer"`
	References []AiReference `json:"references"`
}

// AiReference AI 回答引用
type AiReference struct {
	ActivityID int64  `json:"activityId"`
	Timestamp  int64  `json:"timestamp"`
	Excerpt    string `json:"excerpt"`
}

// ChatMessage AI 助手对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AssistantReply AI 助手回复
type AssistantReply struct {
	Reply     string     `json:"reply"`
	ToolCalls []ToolCall `json:"toolCalls"`
}

// ToolCall AI 工具调用
type ToolCall struct {
	Tool   string      `json:"tool"`
	Input  interface{} `json:"input"`
	Output string      `json:"output"`
}

// AppCategoryInfo 应用分类信息
type AppCategoryInfo struct {
	AppName       string `json:"appName"`
	Category      string `json:"category"`
	IsCustomRule  bool   `json:"isCustomRule"`
	TotalDuration int64  `json:"totalDuration"`
	LastSeen      int64  `json:"lastSeen"`
}

// StorageStats 存储统计
type StorageStats struct {
	ActivityCount      int64  `json:"activityCount"`
	ScreenshotCount    int64  `json:"screenshotCount"`
	DiskUsageMB        int64  `json:"diskUsageMb"`
	MaxStorageMB       int64  `json:"maxStorageMb"`
	OldestActivityDate string `json:"oldestActivityDate"`
	RetentionDays      int    `json:"retentionDays"`
}

// CleanupResult 数据清理结果
type CleanupResult struct {
	DeletedActivities   int64 `json:"deletedActivities"`
	DeletedScreenshots  int64 `json:"deletedScreenshots"`
	FreedMB             int64 `json:"freedMb"`
}

// BatchResult 批量上报结果
type BatchResult struct {
	Total        int `json:"total"`
	Inserted     int `json:"inserted"`
	Deduplicated int `json:"deduplicated"`
}

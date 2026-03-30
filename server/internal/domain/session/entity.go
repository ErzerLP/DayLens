// Package session 工作会话领域实体。
package session

import "daylens-server/internal/domain/activity"

// WorkSession 工作会话
type WorkSession struct {
	StartTime     int64             `json:"startTime"`
	EndTime       int64             `json:"endTime"`
	TotalDuration int64             `json:"totalDuration"`
	Activities    []SessionActivity `json:"activities"`
	DominantApp   string            `json:"dominantApp"`
	Intent        IntentInfo        `json:"intent"`
}

// SessionActivity 会话内的活动摘要
type SessionActivity struct {
	AppName  string `json:"appName"`
	Duration int64  `json:"duration"`
	Title    string `json:"title"`
}

// IntentInfo 意图信息
type IntentInfo struct {
	Label      string   `json:"label"`
	Confidence int      `json:"confidence"`
	Evidence   []string `json:"evidence"`
}

// IntentItem 意图分析条目
type IntentItem struct {
	Label         string `json:"label"`
	TotalDuration int64  `json:"totalDuration"`
	SessionCount  int    `json:"sessionCount"`
	Percentage    int    `json:"percentage"`
}

// IntentAnalysisResult 意图分析结果
type IntentAnalysisResult struct {
	Items                 []IntentItem `json:"items"`
	DominantIntent        string       `json:"dominantIntent"`
	TotalAnalyzedDuration int64        `json:"totalAnalyzedDuration"`
}

// TodoItem 待办事项
type TodoItem struct {
	Title       string `json:"title"`
	Source      string `json:"source"`
	Confidence  string `json:"confidence"`
	ExtractedAt int64  `json:"extractedAt"`
}

// TodoExtractionResult 待办提取结果
type TodoExtractionResult struct {
	Items   []TodoItem `json:"items"`
	Summary string     `json:"summary"`
}

// WeeklyReview 周报
type WeeklyReview struct {
	Period              string              `json:"period"`
	TotalWorkDuration   int64               `json:"totalWorkDuration"`
	AvgDailyDuration    int64               `json:"avgDailyDuration"`
	Content             string              `json:"content"`
	DeepWorkSessions    []DeepWorkSession   `json:"deepWorkSessions"`
	TopApps             []activity.AppUsage `json:"topApps"`
	IntentDistribution  []IntentItem        `json:"intentDistribution"`
}

// DeepWorkSession 深度工作会话
type DeepWorkSession struct {
	Date     string `json:"date"`
	Duration int64  `json:"duration"`
	Focus    string `json:"focus"`
}

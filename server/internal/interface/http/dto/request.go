// Package dto 定义 HTTP 请求数据传输对象。
package dto

// IngestRequest 单条活动上报请求
type IngestRequest struct {
	ClientID           string  `json:"clientId" binding:"required"`
	ClientTs           int64   `json:"clientTs" binding:"required"`
	Timestamp          int64   `json:"timestamp" binding:"required"`
	AppName            string  `json:"appName" binding:"required"`
	WindowTitle        string  `json:"windowTitle" binding:"required"`
	Category           string  `json:"category" binding:"required"`
	SemanticCategory   *string `json:"semanticCategory"`
	SemanticConfidence *int    `json:"semanticConfidence"`
	Duration           int     `json:"duration" binding:"required"`
	BrowserURL         *string `json:"browserUrl"`
	ExecutablePath     *string `json:"executablePath"`
	OcrText            *string `json:"ocrText"`
	ScreenshotKey      *string `json:"screenshotKey"`
}

// BatchRequest 批量活动上报请求
type BatchRequest struct {
	Activities []IngestRequest `json:"activities" binding:"required,max=100"`
}

// GenerateReportRequest 生成日报请求
type GenerateReportRequest struct {
	Date            string `json:"date" binding:"required"`
	ForceRegenerate bool   `json:"forceRegenerate"`
}

// AskRequest AI 问答请求
type AskRequest struct {
	Question string `json:"question" binding:"required"`
	Context  string `json:"context"`
}

// ChatRequest AI 助手对话请求
type ChatRequest struct {
	Messages []ChatMessageDTO `json:"messages" binding:"required"`
	Tools    []string         `json:"tools"`
}

// ChatMessageDTO 对话消息
type ChatMessageDTO struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// SetRuleRequest 设置分类规则请求
type SetRuleRequest struct {
	AppName  string `json:"appName" binding:"required"`
	Category string `json:"category" binding:"required"`
}

// ReclassifyRequest 重新分类请求
type ReclassifyRequest struct {
	AppName     string `json:"appName" binding:"required"`
	NewCategory string `json:"newCategory" binding:"required"`
}

// WeeklyRequest 周报生成请求
type WeeklyRequest struct {
	From string `json:"from" binding:"required"`
	To   string `json:"to" binding:"required"`
}

// SaveAIConfigRequest 保存 AI 配置请求
type SaveAIConfigRequest struct {
	Provider     string `json:"provider" binding:"required"`
	Endpoint     string `json:"endpoint"`
	Model        string `json:"model"`
	APIKey       string `json:"apiKey"`
	CustomPrompt string `json:"customPrompt"`
}

// TestAIRequest 测试 AI 连接请求
type TestAIRequest struct {
	Provider string `json:"provider" binding:"required"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
	APIKey   string `json:"apiKey"`
}

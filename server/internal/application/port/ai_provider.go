package port

import "context"

// Message AI 对话消息
type Message struct {
	Role    string
	Content string
}

// AIProvider AI 服务调用端口
type AIProvider interface {
	// Name 返回提供商标识（如 "ollama"、"openai"）
	Name() string
	// Chat 发送对话请求，返回 AI 回复文本
	Chat(ctx context.Context, messages []Message) (string, error)
	// IsAvailable 检查 AI 服务是否可用
	IsAvailable(ctx context.Context) bool
}

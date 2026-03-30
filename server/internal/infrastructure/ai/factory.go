// Package ai AI 提供商适配器工厂。
package ai

import (
	"daylens-server/config"
	"daylens-server/internal/application/port"
)

// NewProvider 根据配置创建 AI 提供商
func NewProvider(cfg *config.AIConfig) port.AIProvider {
	switch cfg.Provider {
	case "ollama":
		return NewOllama(cfg.Endpoint, cfg.Model)
	case "claude":
		return NewClaude(cfg.Endpoint, cfg.APIKey, cfg.Model)
	case "gemini":
		return NewGemini(cfg.APIKey, cfg.Model)
	default:
		// openai / deepseek / moonshot / siliconflow / 其他兼容 API
		return NewOpenAICompat(cfg.Provider, cfg.Endpoint, cfg.APIKey, cfg.Model)
	}
}

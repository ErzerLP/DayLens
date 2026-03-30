// Package ai AI 提供商适配器工厂。
package ai

import (
	"daylens-server/config"
	"daylens-server/internal/application/port"
)

// 常见 OpenAI 兼容 API 的默认 endpoint
var knownEndpoints = map[string]string{
	"openai":      "https://api.openai.com/v1",
	"deepseek":    "https://api.deepseek.com/v1",
	"moonshot":    "https://api.moonshot.cn/v1",
	"siliconflow": "https://api.siliconflow.cn/v1",
	"groq":        "https://api.groq.com/openai/v1",
	"zhipu":       "https://open.bigmodel.cn/api/paas/v4",
}

// 常见 provider 的默认模型
var knownModels = map[string]string{
	"openai":      "gpt-4o-mini",
	"deepseek":    "deepseek-chat",
	"moonshot":    "moonshot-v1-8k",
	"siliconflow": "Qwen/Qwen2.5-7B-Instruct",
	"groq":        "llama-3.3-70b-versatile",
	"zhipu":       "glm-4-flash",
	"ollama":      "qwen2.5",
	"claude":      "claude-3-5-sonnet-20241022",
	"gemini":      "gemini-2.0-flash",
}

// NewProvider 根据配置创建 AI 提供商
//
// 支持两大类：
//   - Ollama 本地模型（provider = "ollama"）
//   - OpenAI 兼容 API（provider = "openai" / "deepseek" / "moonshot" 等）
//   - Claude Messages API（provider = "claude"）
//   - Gemini API（provider = "gemini"）
//
// 对于 OpenAI 兼容 API，若未填 endpoint 则自动使用预设值。
func NewProvider(cfg *config.AIConfig) port.AIProvider {
	// 自动填充默认 endpoint
	endpoint := cfg.Endpoint
	if endpoint == "" {
		if ep, ok := knownEndpoints[cfg.Provider]; ok {
			endpoint = ep
		}
	}

	// 自动填充默认 model
	model := cfg.Model
	if model == "" {
		if m, ok := knownModels[cfg.Provider]; ok {
			model = m
		}
	}

	switch cfg.Provider {
	case "ollama":
		return NewOllama(endpoint, model, cfg.APIKey)
	case "claude":
		return NewClaude(endpoint, cfg.APIKey, model)
	case "gemini":
		return NewGemini(cfg.APIKey, model)
	default:
		// openai / deepseek / moonshot / siliconflow / groq / zhipu / 任意兼容 API
		return NewOpenAICompat(cfg.Provider, endpoint, cfg.APIKey, model)
	}
}

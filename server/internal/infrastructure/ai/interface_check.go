package ai

import "daylens-server/internal/application/port"

// 编译期接口满足检查
var (
	_ port.AIProvider = (*OllamaProvider)(nil)
	_ port.AIProvider = (*OpenAICompatProvider)(nil)
	_ port.AIProvider = (*ClaudeProvider)(nil)
	_ port.AIProvider = (*GeminiProvider)(nil)
)

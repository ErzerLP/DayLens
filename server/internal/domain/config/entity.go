// Package domainconfig 配置领域实体。
package domainconfig

// AIConfig AI 配置信息（返回给客户端查询）
type AIConfig struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	CustomPrompt string `json:"customPrompt"`
}

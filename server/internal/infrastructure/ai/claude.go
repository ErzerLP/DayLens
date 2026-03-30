package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"daylens-server/internal/application/port"
)

// ClaudeProvider Claude Messages API 适配器
type ClaudeProvider struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

// NewClaude 创建 Claude 提供商
func NewClaude(endpoint, apiKey, model string) *ClaudeProvider {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &ClaudeProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 2 * time.Minute},
	}
}

func (p *ClaudeProvider) Name() string { return "claude" }

// Chat 发送 /v1/messages 请求
func (p *ClaudeProvider) Chat(ctx context.Context, messages []port.Message) (string, error) {
	type claudeMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type claudeReq struct {
		Model     string      `json:"model"`
		MaxTokens int         `json:"max_tokens"`
		System    string      `json:"system,omitempty"`
		Messages  []claudeMsg `json:"messages"`
	}

	// 提取 system message
	var systemPrompt string
	var chatMsgs []claudeMsg
	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt = m.Content
			continue
		}
		chatMsgs = append(chatMsgs, claudeMsg{Role: m.Role, Content: m.Content})
	}

	body, _ := json.Marshal(claudeReq{
		Model:     p.model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  chatMsgs,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("claude request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("claude status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("claude decode: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("claude: no content returned")
	}
	return result.Content[0].Text, nil
}

// IsAvailable 检查 API Key 是否配置
func (p *ClaudeProvider) IsAvailable(_ context.Context) bool {
	return p.apiKey != ""
}

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"daylens-server/internal/application/port"
)

// OpenAICompatProvider OpenAI 兼容 API 适配器（OpenAI/DeepSeek/Moonshot/SiliconFlow/通义）
type OpenAICompatProvider struct {
	name     string
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

// NewOpenAICompat 创建 OpenAI 兼容提供商
func NewOpenAICompat(name, endpoint, apiKey, model string) *OpenAICompatProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAICompatProvider{
		name:     name,
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 2 * time.Minute},
	}
}

func (p *OpenAICompatProvider) Name() string { return p.name }

// Chat 发送 /chat/completions 请求
func (p *OpenAICompatProvider) Chat(ctx context.Context, messages []port.Message) (string, error) {
	type chatMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type chatReq struct {
		Model    string    `json:"model"`
		Messages []chatMsg `json:"messages"`
	}

	msgs := make([]chatMsg, len(messages))
	for i, m := range messages {
		msgs[i] = chatMsg{Role: m.Role, Content: m.Content}
	}

	body, _ := json.Marshal(chatReq{Model: p.model, Messages: msgs})

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("openai decode: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}
	return result.Choices[0].Message.Content, nil
}

// IsAvailable 检查 API Key 是否配置
func (p *OpenAICompatProvider) IsAvailable(_ context.Context) bool {
	return p.apiKey != ""
}

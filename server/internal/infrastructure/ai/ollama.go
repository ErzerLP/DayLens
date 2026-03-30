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

// OllamaProvider Ollama 本地 AI 适配器
//
// 优先走 Ollama 原生 /api/chat 接口；
// 如果配置了 apiKey，则说明走的是代理（如 LiteLLM），自动切换 OpenAI 兼容模式。
type OllamaProvider struct {
	endpoint string
	model    string
	apiKey   string
	client   *http.Client
}

// NewOllama 创建 Ollama 提供商
func NewOllama(endpoint, model, apiKey string) *OllamaProvider {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "qwen2.5"
	}
	return &OllamaProvider{
		endpoint: strings.TrimRight(endpoint, "/"),
		model:    model,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 2 * time.Minute},
	}
}

func (p *OllamaProvider) Name() string { return "ollama (" + p.model + ")" }

// Chat 发送对话请求
func (p *OllamaProvider) Chat(ctx context.Context, messages []port.Message) (string, error) {
	// 有 apiKey 时走 OpenAI 兼容模式（/v1/chat/completions）
	if p.apiKey != "" {
		return p.chatOpenAICompat(ctx, messages)
	}
	return p.chatNative(ctx, messages)
}

// chatNative Ollama 原生接口
func (p *OllamaProvider) chatNative(ctx context.Context, messages []port.Message) (string, error) {
	type ollamaMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type ollamaReq struct {
		Model    string      `json:"model"`
		Messages []ollamaMsg `json:"messages"`
		Stream   bool        `json:"stream"`
	}

	msgs := make([]ollamaMsg, len(messages))
	for i, m := range messages {
		msgs[i] = ollamaMsg{Role: m.Role, Content: m.Content}
	}

	body, _ := json.Marshal(ollamaReq{Model: p.model, Messages: msgs, Stream: false})

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama decode: %w", err)
	}
	return result.Message.Content, nil
}

// chatOpenAICompat Ollama 的 OpenAI 兼容接口（/v1/chat/completions）
func (p *OllamaProvider) chatOpenAICompat(ctx context.Context, messages []port.Message) (string, error) {
	compat := NewOpenAICompat("ollama", p.endpoint+"/v1", p.apiKey, p.model)
	return compat.Chat(ctx, messages)
}

// IsAvailable 检查 Ollama 是否可达
func (p *OllamaProvider) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

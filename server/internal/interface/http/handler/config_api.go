package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"daylens-server/config"
	"daylens-server/internal/application/port"
	"daylens-server/internal/infrastructure/ai"
	"daylens-server/internal/interface/http/dto"
)

// ConfigHandler AI 配置路由处理
type ConfigHandler struct {
	cfg        *config.Config
	aiSwapper  func(provider port.AIProvider) // 回调：热替换 AI provider
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(cfg *config.Config, aiSwapper func(port.AIProvider)) *ConfigHandler {
	return &ConfigHandler{cfg: cfg, aiSwapper: aiSwapper}
}

// GetAI 获取 AI 配置
func (h *ConfigHandler) GetAI(c *gin.Context) {
	// 返回当前配置（隐藏完整 API Key）
	maskedKey := h.cfg.AI.APIKey
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:4] + "****" + maskedKey[len(maskedKey)-4:]
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"provider":     h.cfg.AI.Provider,
		"endpoint":     h.cfg.AI.Endpoint,
		"model":        h.cfg.AI.Model,
		"apiKey":       maskedKey,
		"customPrompt": h.cfg.AI.CustomPrompt,
	}))
}

// SaveAI 保存 AI 配置
func (h *ConfigHandler) SaveAI(c *gin.Context) {
	var req dto.SaveAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}

	// 更新内存配置
	h.cfg.AI.Provider = req.Provider
	if req.Endpoint != "" {
		h.cfg.AI.Endpoint = req.Endpoint
	}
	if req.Model != "" {
		h.cfg.AI.Model = req.Model
	}
	// 只在客户端发送了非掩码的 key 时才更新
	if req.APIKey != "" && (len(req.APIKey) < 8 || req.APIKey[4:8] != "****") {
		h.cfg.AI.APIKey = req.APIKey
	}
	h.cfg.AI.CustomPrompt = req.CustomPrompt

	// 热替换 AI provider
	if h.aiSwapper != nil {
		newProvider := ai.NewProvider(&h.cfg.AI)
		h.aiSwapper(newProvider)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// TestAI 测试 AI 连接
func (h *ConfigHandler) TestAI(c *gin.Context) {
	var req dto.TestAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}

	// 用请求参数临时创建 provider 进行测试
	testCfg := &config.AIConfig{
		Provider: req.Provider,
		Endpoint: req.Endpoint,
		Model:    req.Model,
		APIKey:   req.APIKey,
	}
	// 如果 key 是掩码的，使用当前已保存的 key
	if req.APIKey != "" && len(req.APIKey) >= 8 && req.APIKey[4:8] == "****" {
		testCfg.APIKey = h.cfg.AI.APIKey
	}

	provider := ai.NewProvider(testCfg)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	start := time.Now()
	reply, err := provider.Chat(ctx, []port.Message{
		{Role: "user", Content: "Say hi in one word."},
	})
	latency := time.Since(start).Milliseconds()

	if err != nil {
		c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
			"success":   false,
			"latencyMs": latency,
			"error":     err.Error(),
		}))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"success":   true,
		"latencyMs": latency,
		"reply":     reply,
		"error":     "",
	}))
}

// ListProviders 获取 AI 提供商列表
func (h *ConfigHandler) ListProviders(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"items": []gin.H{
			{
				"id":              "gemini",
				"name":            "Google Gemini",
				"defaultEndpoint": "",
				"defaultModel":    "gemini-2.0-flash",
				"requiresApiKey":  true,
			},
			{
				"id":              "openai",
				"name":            "OpenAI / 兼容API",
				"defaultEndpoint": "https://api.openai.com/v1",
				"defaultModel":    "gpt-4o-mini",
				"requiresApiKey":  true,
			},
			{
				"id":              "deepseek",
				"name":            "DeepSeek",
				"defaultEndpoint": "https://api.deepseek.com/v1",
				"defaultModel":    "deepseek-chat",
				"requiresApiKey":  true,
			},
			{
				"id":              "claude",
				"name":            "Anthropic Claude",
				"defaultEndpoint": "https://api.anthropic.com",
				"defaultModel":    "claude-3-5-sonnet-20241022",
				"requiresApiKey":  true,
			},
			{
				"id":              "ollama",
				"name":            "Ollama (本地)",
				"defaultEndpoint": "http://localhost:11434",
				"defaultModel":    "qwen2.5",
				"requiresApiKey":  false,
			},
			{
				"id":              "siliconflow",
				"name":            "SiliconFlow 硅基流动",
				"defaultEndpoint": "https://api.siliconflow.cn/v1",
				"defaultModel":    "Qwen/Qwen2.5-7B-Instruct",
				"requiresApiKey":  true,
			},
			{
				"id":              "groq",
				"name":            "Groq",
				"defaultEndpoint": "https://api.groq.com/openai/v1",
				"defaultModel":    "llama-3.3-70b-versatile",
				"requiresApiKey":  true,
			},
		},
	}))
}

// GetTimezone 获取当前时区
func (h *ConfigHandler) GetTimezone(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"timezone": h.cfg.Server.Timezone,
	}))
}

// SaveTimezone 保存时区
func (h *ConfigHandler) SaveTimezone(c *gin.Context) {
	var req struct {
		Timezone string `json:"timezone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}
	// 验证时区有效性
	if _, err := time.LoadLocation(req.Timezone); err != nil {
		c.JSON(http.StatusOK, dto.ErrorResponse(42201, "无效的时区: "+req.Timezone))
		return
	}
	h.cfg.Server.Timezone = req.Timezone
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}


package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/interface/http/dto"
)

// ConfigHandler AI 配置路由处理
type ConfigHandler struct{}

// NewConfigHandler 创建配置处理器
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// GetAI 获取 AI 配置
func (h *ConfigHandler) GetAI(c *gin.Context) {
	// TODO: 阶段 8 实现 AI Provider 后接入
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"provider": "ollama",
		"endpoint": "http://localhost:11434",
		"model":    "qwen2.5",
		"apiKey":   "",
	}))
}

// SaveAI 保存 AI 配置
func (h *ConfigHandler) SaveAI(c *gin.Context) {
	var req dto.SaveAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}
	// TODO: 阶段 8
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// TestAI 测试 AI 连接
func (h *ConfigHandler) TestAI(c *gin.Context) {
	var req dto.TestAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}
	// TODO: 阶段 8
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"success":       false,
		"latencyMs":     0,
		"error":         "AI Provider 尚未实现",
	}))
}

// ListProviders 获取 AI 提供商列表
func (h *ConfigHandler) ListProviders(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"items": []gin.H{
			{
				"id":                 "ollama",
				"name":               "Ollama (本地)",
				"defaultEndpoint":    "http://localhost:11434",
				"defaultModel":       "qwen2.5",
				"requiresApiKey":     false,
				"isOpenaiCompatible": false,
			},
			{
				"id":                 "openai",
				"name":               "OpenAI / 兼容API",
				"defaultEndpoint":    "https://api.openai.com/v1",
				"defaultModel":       "gpt-4o-mini",
				"requiresApiKey":     true,
				"isOpenaiCompatible": true,
			},
		},
	}))
}

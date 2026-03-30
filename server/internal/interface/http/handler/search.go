package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/domain/activity"
	"daylens-server/internal/interface/http/dto"
)

// SearchHandler 搜索与 AI 路由处理
type SearchHandler struct {
	svc *application.SearchService
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(svc *application.SearchService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Search 全文搜索
func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	limit := queryInt(c, "limit", 20)
	items, total, err := h.svc.Search(c.Request.Context(), defaultUserID, q, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"items": items, "total": total}))
}

// Ask AI 问答
func (h *SearchHandler) Ask(c *gin.Context) {
	var req dto.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}
	answer, err := h.svc.Ask(c.Request.Context(), defaultUserID, req.Question, req.Context)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(answer))
}

// Chat AI 助手对话
func (h *SearchHandler) Chat(c *gin.Context) {
	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}

	messages := make([]activity.ChatMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = activity.ChatMessage{Role: m.Role, Content: m.Content}
	}

	reply, err := h.svc.Chat(c.Request.Context(), defaultUserID, messages, req.Tools)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(reply))
}

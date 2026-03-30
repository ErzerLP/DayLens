package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/interface/http/dto"
)

// SessionHandler 工作会话路由处理
type SessionHandler struct {
	svc *application.SessionService
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler(svc *application.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

// GetSessions 获取工作会话列表
func (h *SessionHandler) GetSessions(c *gin.Context) {
	date := queryDate(c)
	items, err := h.svc.GetSessions(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"items": items}))
}

// GetIntents 获取意图分析
func (h *SessionHandler) GetIntents(c *gin.Context) {
	date := queryDate(c)
	result, err := h.svc.GetIntents(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

// GetTodos 获取待办事项
func (h *SessionHandler) GetTodos(c *gin.Context) {
	from := c.DefaultQuery("from", queryDate(c))
	to := c.DefaultQuery("to", queryDate(c))
	result, err := h.svc.GetTodos(c.Request.Context(), defaultUserID, from, to)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

// WeeklyReview 生成周报
func (h *SessionHandler) WeeklyReview(c *gin.Context) {
	var req dto.WeeklyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}
	result, err := h.svc.GenerateWeekly(c.Request.Context(), defaultUserID, req.From, req.To)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

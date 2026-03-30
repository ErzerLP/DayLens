package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/interface/http/dto"
)

// InternalHandler 内部 API 路由处理（无认证）
type InternalHandler struct {
	activitySvc *application.ActivityService
	reportSvc   *application.ReportService
	sessionSvc  *application.SessionService
	searchSvc   *application.SearchService
}

// NewInternalHandler 创建内部 API 处理器
func NewInternalHandler(
	activitySvc *application.ActivityService,
	reportSvc *application.ReportService,
	sessionSvc *application.SessionService,
	searchSvc *application.SearchService,
) *InternalHandler {
	return &InternalHandler{
		activitySvc: activitySvc,
		reportSvc:   reportSvc,
		sessionSvc:  sessionSvc,
		searchSvc:   searchSvc,
	}
}

// Health 健康检查
func (h *InternalHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"status": "ok"}))
}

// Stats 获取统计（内部）
func (h *InternalHandler) Stats(c *gin.Context) {
	date := queryDate(c)
	stats, err := h.activitySvc.GetStats(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(stats))
}

// Activities 获取活动列表（内部）
func (h *InternalHandler) Activities(c *gin.Context) {
	date := queryDate(c)
	limit := queryInt(c, "limit", 50)
	items, total, err := h.activitySvc.GetTimeline(c.Request.Context(), defaultUserID, date, "", "", limit, 0)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse(items, total, limit, 0))
}

// Sessions 获取工作会话（内部）
func (h *InternalHandler) Sessions(c *gin.Context) {
	date := queryDate(c)
	items, err := h.sessionSvc.GetSessions(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"items": items}))
}

// Intents 获取意图分析（内部）
func (h *InternalHandler) Intents(c *gin.Context) {
	date := queryDate(c)
	result, err := h.sessionSvc.GetIntents(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

// Report 获取日报（内部）
func (h *InternalHandler) Report(c *gin.Context) {
	date := c.Param("date")
	r, err := h.reportSvc.GetByDate(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(r))
}

// WeeklyReview 获取周报（内部）
func (h *InternalHandler) WeeklyReview(c *gin.Context) {
	from := c.DefaultQuery("from", queryDate(c))
	to := c.DefaultQuery("to", queryDate(c))
	result, err := h.sessionSvc.GenerateWeekly(c.Request.Context(), defaultUserID, from, to)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

// Todos 获取待办事项（内部）
func (h *InternalHandler) Todos(c *gin.Context) {
	from := c.DefaultQuery("from", queryDate(c))
	to := c.DefaultQuery("to", queryDate(c))
	result, err := h.sessionSvc.GetTodos(c.Request.Context(), defaultUserID, from, to)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

// Search 搜索（内部）
func (h *InternalHandler) Search(c *gin.Context) {
	q := c.Query("q")
	limit := queryInt(c, "limit", 20)
	items, total, err := h.searchSvc.Search(c.Request.Context(), defaultUserID, q, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"items": items, "total": total}))
}

// Export 导出日报（内部）
func (h *InternalHandler) Export(c *gin.Context) {
	date := queryDate(c)
	content, err := h.reportSvc.ExportMarkdown(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(content))
}

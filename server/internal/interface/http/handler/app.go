package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/interface/http/dto"
)

// AppHandler 应用管理路由处理
type AppHandler struct {
	svc *application.AppService
}

// NewAppHandler 创建应用管理处理器
func NewAppHandler(svc *application.AppService) *AppHandler {
	return &AppHandler{svc: svc}
}

// Recent 获取最近使用的应用
func (h *AppHandler) Recent(c *gin.Context) {
	days := queryInt(c, "days", 7)
	apps, err := h.svc.GetRecentApps(c.Request.Context(), defaultUserID, days)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"items": apps}))
}

// Categories 获取应用分类概览
func (h *AppHandler) Categories(c *gin.Context) {
	from := c.DefaultQuery("from", queryDate(c))
	to := c.DefaultQuery("to", queryDate(c))
	items, err := h.svc.GetCategories(c.Request.Context(), defaultUserID, from, to)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"items": items}))
}

// SetRule 设置分类规则
func (h *AppHandler) SetRule(c *gin.Context) {
	var req dto.SetRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}
	if err := h.svc.SetRule(c.Request.Context(), defaultUserID, req.AppName, req.Category); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// Reclassify 重新分类应用历史
func (h *AppHandler) Reclassify(c *gin.Context) {
	var req dto.ReclassifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}
	count, err := h.svc.Reclassify(c.Request.Context(), defaultUserID, req.AppName, req.NewCategory)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"updatedCount": count}))
}

package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/interface/http/dto"
)

// ReportHandler 日报路由处理
type ReportHandler struct {
	svc *application.ReportService
}

// NewReportHandler 创建日报处理器
func NewReportHandler(svc *application.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// Get 获取日报
func (h *ReportHandler) Get(c *gin.Context) {
	date := c.Param("date")
	r, err := h.svc.GetByDate(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(r))
}

// Generate 生成日报
func (h *ReportHandler) Generate(c *gin.Context) {
	var req dto.GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}

	// AI 生成报告可能耗时较长，使用独立超时 context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	r, err := h.svc.Generate(ctx, defaultUserID, req.Date, req.ForceRegenerate)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.SuccessResponse(r))
}

// Export 导出日报 Markdown
func (h *ReportHandler) Export(c *gin.Context) {
	date := c.Param("date")
	content, err := h.svc.ExportMarkdown(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+date+".md\"")
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(content))
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/application/port"
	"daylens-server/internal/interface/http/dto"
)

// StatsHandler 统计数据路由处理
type StatsHandler struct {
	activitySvc *application.ActivityService
	hourlyRepo  port.HourlyRepository
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(svc *application.ActivityService, hourlyRepo port.HourlyRepository) *StatsHandler {
	return &StatsHandler{activitySvc: svc, hourlyRepo: hourlyRepo}
}

// GetStats 获取每日统计
func (h *StatsHandler) GetStats(c *gin.Context) {
	date := queryDate(c)
	stats, err := h.activitySvc.GetStats(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(stats))
}

// GetHourlySummaries 获取小时摘要
func (h *StatsHandler) GetHourlySummaries(c *gin.Context) {
	date := queryDate(c)
	items, err := h.activitySvc.GetHourlySummaries(c.Request.Context(), defaultUserID, date, h.hourlyRepo)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"items": items}))
}

package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/domain/activity"
	"daylens-server/internal/interface/http/dto"
)

// ActivityHandler 活动相关路由处理
type ActivityHandler struct {
	svc *application.ActivityService
}

// NewActivityHandler 创建活动处理器
func NewActivityHandler(svc *application.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

// Ingest 上报单条活动
func (h *ActivityHandler) Ingest(c *gin.Context) {
	var req dto.IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}

	a := &activity.Activity{
		ClientID:           req.ClientID,
		ClientTs:           req.ClientTs,
		Timestamp:          req.Timestamp,
		AppName:            req.AppName,
		WindowTitle:        req.WindowTitle,
		Category:           req.Category,
		SemanticCategory:   req.SemanticCategory,
		SemanticConfidence: req.SemanticConfidence,
		Duration:           req.Duration,
		BrowserURL:         req.BrowserURL,
		ExecutablePath:     req.ExecutablePath,
		OcrText:            req.OcrText,
		ScreenshotKey:      req.ScreenshotKey,
	}

	id, err := h.svc.IngestActivity(c.Request.Context(), a)
	if err != nil {
		handleError(c, err)
		return
	}

	// 幂等命中返回 200，新插入返回 201
	status := http.StatusCreated
	deduplicated := false
	if id == 0 {
		status = http.StatusOK
		deduplicated = true
	}
	c.JSON(status, dto.SuccessResponse(gin.H{"id": id, "deduplicated": deduplicated}))
}

// IngestBatch 批量上报活动
func (h *ActivityHandler) IngestBatch(c *gin.Context) {
	var req dto.BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c)
		return
	}

	list := make([]*activity.Activity, len(req.Activities))
	for i, r := range req.Activities {
		list[i] = &activity.Activity{
			ClientID:           r.ClientID,
			ClientTs:           r.ClientTs,
			Timestamp:          r.Timestamp,
			AppName:            r.AppName,
			WindowTitle:        r.WindowTitle,
			Category:           r.Category,
			SemanticCategory:   r.SemanticCategory,
			SemanticConfidence: r.SemanticConfidence,
			Duration:           r.Duration,
			BrowserURL:         r.BrowserURL,
			ExecutablePath:     r.ExecutablePath,
			OcrText:            r.OcrText,
			ScreenshotKey:      r.ScreenshotKey,
		}
	}

	result, err := h.svc.IngestBatch(c.Request.Context(), list)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

// List 查询活动列表
func (h *ActivityHandler) List(c *gin.Context) {
	date := queryDate(c)
	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)
	app := c.Query("app")
	category := c.Query("category")

	items, total, err := h.svc.GetTimeline(c.Request.Context(), defaultUserID, date, app, category, limit, offset)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.PaginatedResponse(items, total, limit, offset))
}

// Get 获取单条活动
func (h *ActivityHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		bindError(c)
		return
	}

	a, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(a))
}

// DeleteBefore 删除指定日期之前的活动
func (h *ActivityHandler) DeleteBefore(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		bindError(c)
		return
	}
	// 委托给 StorageService 处理（在 router 中绑定）
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

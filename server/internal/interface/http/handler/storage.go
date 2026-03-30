package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/application"
	"daylens-server/internal/interface/http/dto"
)

// StorageHandler 存储管理路由处理
type StorageHandler struct {
	svc *application.StorageService
}

// NewStorageHandler 创建存储处理器
func NewStorageHandler(svc *application.StorageService) *StorageHandler {
	return &StorageHandler{svc: svc}
}

// Stats 获取存储统计
func (h *StorageHandler) Stats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(stats))
}

// DeleteBefore 清理指定日期之前的数据
func (h *StorageHandler) DeleteBefore(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		bindError(c)
		return
	}
	result, err := h.svc.DeleteBefore(c.Request.Context(), defaultUserID, date)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

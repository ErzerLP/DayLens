package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"daylens-server/internal/application/port"
	"daylens-server/internal/interface/http/dto"
)

// ScreenshotHandler 截图上传与获取路由处理
type ScreenshotHandler struct {
	storage port.FileStorage
}

// NewScreenshotHandler 创建截图处理器
func NewScreenshotHandler(storage port.FileStorage) *ScreenshotHandler {
	return &ScreenshotHandler{storage: storage}
}

// Upload 上传截图
func (h *ScreenshotHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		bindError(c)
		return
	}
	defer file.Close()

	// 校验大小（最大 5MB）
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse(42201, "文件超过 5MB 限制"))
		return
	}

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(50001, "读取文件失败"))
		return
	}

	// 生成存储 key: 年/月/日/uuid.ext
	ext := ".jpg"
	ct := header.Header.Get("Content-Type")
	if ct == "image/png" {
		ext = ".png"
	}
	now := time.Now()
	key := fmt.Sprintf("%d/%02d/%02d/%s%s", now.Year(), now.Month(), now.Day(), uuid.New().String()[:8], ext)

	if err := h.storage.Save(c.Request.Context(), key, data); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(50001, "存储截图失败"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(gin.H{"key": key}))
}

// Get 获取截图（支持缩略图）
func (h *ScreenshotHandler) Get(c *gin.Context) {
	key := c.Param("key")
	size := c.DefaultQuery("size", "thumb")

	var data []byte
	var err error

	if size == "full" {
		data, err = h.storage.Get(c.Request.Context(), key)
	} else {
		data, err = h.storage.GetThumbnail(c.Request.Context(), key, 360)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse(40401, "截图不存在"))
		return
	}

	contentType := "image/jpeg"
	if len(key) > 4 && key[len(key)-4:] == ".png" {
		contentType = "image/png"
	}
	c.Data(http.StatusOK, contentType, data)
}

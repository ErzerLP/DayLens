// Package handler HTTP 路由处理器。
package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/interface/http/dto"
	"daylens-server/internal/shared"
)

// handleError 统一错误翻译（error → HTTP 状态码 + 错误码）
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse(shared.CodeNotFound, "资源不存在"))
	case errors.Is(err, shared.ErrDuplicate):
		c.JSON(http.StatusConflict, dto.ErrorResponse(shared.CodeDuplicate, "重复数据"))
	case errors.Is(err, shared.ErrFieldMissing):
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse(shared.CodeFieldMissing, err.Error()))
	case errors.Is(err, shared.ErrFieldInvalid):
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse(shared.CodeFieldInvalid, err.Error()))
	case errors.Is(err, shared.ErrUnavailable):
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse(shared.CodeUnavailable, "服务不可用"))
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(shared.CodeInternal, "内部错误"))
	}
}

// bindError 请求体解析失败响应
func bindError(c *gin.Context) {
	c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse(shared.CodeBadRequest, "请求体解析失败"))
}

// queryDate 从 query 获取 date 参数，默认今天
func queryDate(c *gin.Context) string {
	d := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	return d
}

// queryInt 从 query 获取 int，带默认值
func queryInt(c *gin.Context, key string, defaultVal int) int {
	val := c.DefaultQuery(key, "")
	if val == "" {
		return defaultVal
	}
	var n int
	for _, ch := range val {
		if ch < '0' || ch > '9' {
			return defaultVal
		}
		n = n*10 + int(ch-'0')
	}
	return n
}

// defaultUserID 当前固定 userID=1
const defaultUserID = 1

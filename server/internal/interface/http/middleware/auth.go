// Package middleware HTTP 中间件。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"daylens-server/internal/interface/http/dto"
	"daylens-server/internal/shared"
)

// Auth Bearer Token 认证中间件
func Auth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				dto.ErrorResponse(shared.CodeTokenMissing, "Token 缺失"))
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				dto.ErrorResponse(shared.CodeTokenInvalid, "Token 无效"))
			return
		}

		c.Next()
	}
}

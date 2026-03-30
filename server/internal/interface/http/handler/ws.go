package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"nhooyr.io/websocket"

	"daylens-server/internal/infrastructure/ws"
)

// WSHandler WebSocket 升级处理
type WSHandler struct {
	hub *ws.Hub
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// Upgrade HTTP → WebSocket 协议升级
func (h *WSHandler) Upgrade(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // 允许跨域 WebSocket
	})
	if err != nil {
		slog.Error("WebSocket 升级失败", "error", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client := ws.NewClient(h.hub, conn)
	client.Run(c.Request.Context())
}

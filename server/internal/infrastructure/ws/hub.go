// Package ws WebSocket 连接管理，实现 port.EventBus。
package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
)

// Event WebSocket 推送事件
type Event struct {
	Type string      `json:"event"`
	Data interface{} `json:"data"`
}

// Hub 管理所有 WebSocket 客户端连接和事件广播
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
	}
}

// Register 注册客户端
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
	slog.Info("WebSocket 客户端已连接", "total", len(h.clients))
}

// Unregister 注销客户端
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	slog.Info("WebSocket 客户端已断开", "total", len(h.clients))
}

// Publish 广播事件到所有客户端（实现 port.EventBus）
func (h *Hub) Publish(event string, payload interface{}) {
	msg, err := json.Marshal(&Event{Type: event, Data: payload})
	if err != nil {
		slog.Error("序列化 WebSocket 事件失败", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- msg:
		default:
			// 发送缓冲区已满，跳过该客户端
			slog.Warn("WebSocket 客户端发送缓冲区已满，跳过")
		}
	}

	slog.Debug("WebSocket 事件已广播", "event", event, "clients", len(h.clients))
}

// ClientCount 当前连接数
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

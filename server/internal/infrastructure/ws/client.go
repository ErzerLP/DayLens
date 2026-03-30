package ws

import (
	"context"
	"log/slog"
	"time"

	"nhooyr.io/websocket"
)

const (
	// 发送缓冲区大小
	sendBufSize = 64
	// 写入超时
	writeTimeout = 10 * time.Second
	// 心跳间隔
	pingInterval = 30 * time.Second
)

// Client 单个 WebSocket 连接
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// NewClient 创建客户端
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, sendBufSize),
	}
}

// Run 启动读写协程，阻塞直到连接关闭
func (c *Client) Run(ctx context.Context) {
	c.hub.Register(c)
	defer c.hub.Unregister(c)

	// 写协程
	go c.writePump(ctx)
	// 读协程（阻塞，处理客户端消息 + 检测断开）
	c.readPump(ctx)
}

// readPump 持续读取客户端消息（主要用于检测断开）
func (c *Client) readPump(ctx context.Context) {
	defer c.conn.Close(websocket.StatusNormalClosure, "")
	for {
		_, _, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				slog.Debug("WebSocket 客户端正常断开")
			} else {
				slog.Debug("WebSocket 读取错误", "error", err)
			}
			return
		}
		// 忽略客户端发来的消息（单向推送）
	}
}

// writePump 持续将事件写入 WebSocket
func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return // send channel 已关闭
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				slog.Debug("WebSocket 写入失败", "error", err)
				return
			}

		case <-ticker.C:
			// 心跳 ping
			pingCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				slog.Debug("WebSocket Ping 失败", "error", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

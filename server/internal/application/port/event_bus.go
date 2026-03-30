package port

// EventBus 事件推送端口，解耦应用层与 WebSocket
type EventBus interface {
	// Publish 广播事件到所有已连接客户端
	Publish(event string, payload interface{})
}

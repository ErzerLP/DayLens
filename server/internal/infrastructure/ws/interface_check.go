package ws

import "daylens-server/internal/application/port"

// 编译期接口满足检查
var _ port.EventBus = (*Hub)(nil)

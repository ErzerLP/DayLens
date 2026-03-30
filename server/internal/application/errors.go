package application

import (
	"daylens-server/internal/shared"
)

// 应用层哨兵错误别名，方便各 Service 使用
var (
	ErrFieldMissing = shared.ErrFieldMissing
	ErrFieldInvalid = shared.ErrFieldInvalid
	ErrNotFound     = shared.ErrNotFound
)

package storage

import "daylens-server/internal/application/port"

// 编译期接口满足检查
var _ port.FileStorage = (*LocalFileStorage)(nil)

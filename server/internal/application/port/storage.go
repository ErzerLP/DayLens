package port

import (
	"context"

	"daylens-server/internal/domain/activity"
)

// StorageStats 存储统计信息
type StorageStats = activity.StorageStats

// FileStorage 文件存储端口
type FileStorage interface {
	// Save 保存文件，key 为存储路径（如 "2026/03/29/uuid.jpg"）
	Save(ctx context.Context, key string, data []byte) error
	// Get 获取文件原始内容
	Get(ctx context.Context, key string) ([]byte, error)
	// GetThumbnail 获取缩略图（指定宽度）
	GetThumbnail(ctx context.Context, key string, width int) ([]byte, error)
	// Delete 删除单个文件
	Delete(ctx context.Context, key string) error
	// Stats 获取存储统计
	Stats(ctx context.Context) (*StorageStats, error)
}

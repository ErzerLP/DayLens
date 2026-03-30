package application

import (
	"context"
	"fmt"
	"time"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
)

// StorageService 存储管理用例编排
type StorageService struct {
	activityRepo port.ActivityRepository
	fileStorage  port.FileStorage
}

// NewStorageService 创建存储管理服务
func NewStorageService(activityRepo port.ActivityRepository, fileStorage port.FileStorage) *StorageService {
	return &StorageService{activityRepo: activityRepo, fileStorage: fileStorage}
}

// GetStats 获取存储统计信息
func (s *StorageService) GetStats(ctx context.Context) (*activity.StorageStats, error) {
	stats, err := s.fileStorage.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get storage stats: %w", err)
	}
	return stats, nil
}

// DeleteBefore 删除指定日期之前的活动和截图
func (s *StorageService) DeleteBefore(ctx context.Context, userID int, beforeDate string) (*activity.CleanupResult, error) {
	before, err := time.Parse("2006-01-02", beforeDate)
	if err != nil {
		return nil, fmt.Errorf("delete before: %w: date format", ErrFieldInvalid)
	}

	deletedActivities, err := s.activityRepo.DeleteBefore(ctx, userID, before)
	if err != nil {
		return nil, fmt.Errorf("delete activities: %w", err)
	}

	// TODO: 阶段 6 集成截图清理
	return &activity.CleanupResult{
		DeletedActivities:  deletedActivities,
		DeletedScreenshots: 0,
		FreedMB:            0,
	}, nil
}

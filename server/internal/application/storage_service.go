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

	// 1. 先查询即将删除的活动，提取截图 key
	activities, _, err := s.activityRepo.QueryByDate(ctx, userID, beforeDate, "", "", 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("query activities for cleanup: %w", err)
	}

	// 收集所有截图 key（滤掉 beforeDate 之后的）
	var screenshotKeys []string
	for _, a := range activities {
		ts := time.Unix(a.Timestamp, 0)
		if ts.Before(before) && a.ScreenshotKey != nil && *a.ScreenshotKey != "" {
			screenshotKeys = append(screenshotKeys, *a.ScreenshotKey)
		}
	}

	// 2. 删除截图文件
	deletedScreenshots := int64(0)
	for _, key := range screenshotKeys {
		if err := s.fileStorage.Delete(ctx, key); err == nil {
			deletedScreenshots++
		}
	}

	// 3. 删除活动记录
	deletedActivities, err := s.activityRepo.DeleteBefore(ctx, userID, before)
	if err != nil {
		return nil, fmt.Errorf("delete activities: %w", err)
	}

	return &activity.CleanupResult{
		DeletedActivities:  deletedActivities,
		DeletedScreenshots: deletedScreenshots,
		FreedMB:            0, // 精确计算开销大，暂时忽略
	}, nil
}

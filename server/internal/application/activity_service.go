// Package application 应用层用例编排。
package application

import (
	"context"
	"fmt"
	"log/slog"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
)

// ActivityService 活动数据用例编排
type ActivityService struct {
	repo     port.ActivityRepository
	eventBus port.EventBus
}

// NewActivityService 创建活动服务
func NewActivityService(repo port.ActivityRepository, eventBus port.EventBus) *ActivityService {
	return &ActivityService{repo: repo, eventBus: eventBus}
}

// IngestActivity 接收并存储单条活动记录，幂等去重
func (s *ActivityService) IngestActivity(ctx context.Context, a *activity.Activity) (int64, error) {
	if a.AppName == "" {
		return 0, fmt.Errorf("ingest activity: %w: appName", ErrFieldMissing)
	}

	id, err := s.repo.Insert(ctx, a)
	if err != nil {
		return 0, fmt.Errorf("ingest activity: %w", err)
	}

	// 上报成功后推送事件
	if id > 0 {
		s.eventBus.Publish("activity_received", map[string]interface{}{"id": id})
	}

	return id, nil
}

// IngestBatch 批量接收活动记录
func (s *ActivityService) IngestBatch(ctx context.Context, list []*activity.Activity) (*activity.BatchResult, error) {
	total := len(list)
	if total == 0 {
		return &activity.BatchResult{}, nil
	}
	if total > 100 {
		return nil, fmt.Errorf("ingest batch: 最多 100 条，收到 %d 条", total)
	}

	inserted, err := s.repo.InsertBatch(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("ingest batch: %w", err)
	}

	slog.Info("批量上报完成", "total", total, "inserted", inserted)
	s.eventBus.Publish("batch_received", map[string]interface{}{
		"total": total, "inserted": inserted,
	})

	return &activity.BatchResult{
		Total:        total,
		Inserted:     inserted,
		Deduplicated: total - inserted,
	}, nil
}

// GetTimeline 按日期分页查询活动时间线
func (s *ActivityService) GetTimeline(ctx context.Context, userID int, date, app, category string, limit, offset int) ([]*activity.Activity, int, error) {
	items, total, err := s.repo.QueryByDate(ctx, userID, date, app, category, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get timeline: %w", err)
	}
	return items, total, nil
}

// GetByID 查询单条活动
func (s *ActivityService) GetByID(ctx context.Context, id int64) (*activity.Activity, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get activity %d: %w", id, err)
	}
	return a, nil
}

// GetStats 获取指定日期的统计数据
func (s *ActivityService) GetStats(ctx context.Context, userID int, date string) (*activity.DailyStats, error) {
	stats, err := s.repo.GetDailyStats(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return stats, nil
}

// GetHourlySummaries 获取指定日期的小时摘要
func (s *ActivityService) GetHourlySummaries(ctx context.Context, userID int, date string, hourlyRepo port.HourlyRepository) ([]*activity.HourlySummary, error) {
	items, err := hourlyRepo.GetByDate(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get hourly summaries: %w", err)
	}
	return items, nil
}

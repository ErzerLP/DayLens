package application

import (
	"context"
	"fmt"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
)

// AppService 应用管理与分类规则用例编排
type AppService struct {
	activityRepo port.ActivityRepository
	categoryRepo port.CategoryRuleRepository
}

// NewAppService 创建应用管理服务
func NewAppService(activityRepo port.ActivityRepository, categoryRepo port.CategoryRuleRepository) *AppService {
	return &AppService{activityRepo: activityRepo, categoryRepo: categoryRepo}
}

// GetRecentApps 获取最近 N 天使用过的应用列表
func (s *AppService) GetRecentApps(ctx context.Context, userID int, days int) ([]string, error) {
	apps, err := s.activityRepo.GetRecentApps(ctx, userID, days)
	if err != nil {
		return nil, fmt.Errorf("get recent apps: %w", err)
	}
	return apps, nil
}

// GetCategories 获取应用分类概览
func (s *AppService) GetCategories(ctx context.Context, userID int, from, to string) ([]*activity.AppCategoryInfo, error) {
	items, err := s.activityRepo.GetAppCategories(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	return items, nil
}

// SetRule 设置自定义分类规则
func (s *AppService) SetRule(ctx context.Context, userID int, appName, category string) error {
	if appName == "" || category == "" {
		return fmt.Errorf("set rule: %w", ErrFieldMissing)
	}
	if err := s.categoryRepo.Upsert(ctx, userID, appName, category); err != nil {
		return fmt.Errorf("set rule: %w", err)
	}
	return nil
}

// Reclassify 更新指定应用所有历史记录的分类
func (s *AppService) Reclassify(ctx context.Context, userID int, appName, newCategory string) (int64, error) {
	if appName == "" || newCategory == "" {
		return 0, fmt.Errorf("reclassify: %w", ErrFieldMissing)
	}
	count, err := s.activityRepo.Reclassify(ctx, userID, appName, newCategory)
	if err != nil {
		return 0, fmt.Errorf("reclassify: %w", err)
	}
	return count, nil
}

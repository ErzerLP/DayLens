// Package port 定义应用层与外部世界的交互端口接口。
package port

import (
	"context"
	"time"

	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/report"
)

// ActivityRepository 活动数据持久化端口
type ActivityRepository interface {
	// Insert 插入单条活动，幂等（重复返回 0, nil）
	Insert(ctx context.Context, a *activity.Activity) (int64, error)
	// InsertBatch 批量插入活动，返回实际插入数
	InsertBatch(ctx context.Context, list []*activity.Activity) (inserted int, err error)
	// QueryByDate 按日期分页查询，支持 app/category 筛选
	QueryByDate(ctx context.Context, userID int, date string, app, category string, limit, offset int) ([]*activity.Activity, int, error)
	// GetByID 按 ID 查询单条
	GetByID(ctx context.Context, id int64) (*activity.Activity, error)
	// GetDailyStats 获取指定日期的统计聚合
	GetDailyStats(ctx context.Context, userID int, date string) (*activity.DailyStats, error)
	// Search 全文搜索活动
	Search(ctx context.Context, userID int, query string, limit int) ([]*activity.SearchResultItem, int, error)
	// DeleteBefore 删除指定时间之前的活动，返回删除数
	DeleteBefore(ctx context.Context, userID int, before time.Time) (int64, error)
	// GetRecentApps 获取最近 N 天使用过的应用名列表
	GetRecentApps(ctx context.Context, userID int, days int) ([]string, error)
	// GetAppCategories 获取应用分类概览
	GetAppCategories(ctx context.Context, userID int, from, to string) ([]*activity.AppCategoryInfo, error)
	// Reclassify 更新指定应用的分类，返回更新数
	Reclassify(ctx context.Context, userID int, appName, newCategory string) (int64, error)
}

// ReportRepository 日报持久化端口
type ReportRepository interface {
	// Upsert 插入或更新日报
	Upsert(ctx context.Context, userID int, r *report.DailyReport) error
	// GetByDate 按日期查询日报
	GetByDate(ctx context.Context, userID int, date string) (*report.DailyReport, error)
}

// HourlyRepository 小时摘要持久化端口
type HourlyRepository interface {
	// Upsert 插入或更新小时摘要
	Upsert(ctx context.Context, userID int, s *activity.HourlySummary, date string) error
	// GetByDate 按日期查询全部小时摘要
	GetByDate(ctx context.Context, userID int, date string) ([]*activity.HourlySummary, error)
}

// CategoryRuleRepository 自定义分类规则持久化端口
type CategoryRuleRepository interface {
	// Upsert 插入或更新分类规则
	Upsert(ctx context.Context, userID int, appName, category string) error
	// GetAll 获取用户全部自定义规则
	GetAll(ctx context.Context, userID int) ([]*activity.CategoryRule, error)
}

package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"daylens-server/internal/domain/activity"
)

// PgHourlyRepo 小时摘要 PostgreSQL 仓储
type PgHourlyRepo struct {
	pool *pgxpool.Pool
}

// NewPgHourlyRepo 创建小时摘要仓储
func NewPgHourlyRepo(pool *pgxpool.Pool) *PgHourlyRepo {
	return &PgHourlyRepo{pool: pool}
}

// Upsert 插入或更新小时摘要
func (r *PgHourlyRepo) Upsert(ctx context.Context, userID int, s *activity.HourlySummary, date string) error {
	screenshots := ""
	if len(s.RepresentativeScreenshots) > 0 {
		// 简单 JSON 序列化截图列表
		screenshots = "["
		for i, key := range s.RepresentativeScreenshots {
			if i > 0 {
				screenshots += ","
			}
			screenshots += fmt.Sprintf("%q", key)
		}
		screenshots += "]"
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO hourly_summaries (user_id, date, hour, summary, main_apps,
			activity_count, total_duration, representative_screenshots)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, date, hour) DO UPDATE SET
			summary = EXCLUDED.summary,
			main_apps = EXCLUDED.main_apps,
			activity_count = EXCLUDED.activity_count,
			total_duration = EXCLUDED.total_duration,
			representative_screenshots = EXCLUDED.representative_screenshots`,
		userID, date, s.Hour, s.Summary, s.MainApps,
		s.ActivityCount, s.TotalDuration, screenshots)
	if err != nil {
		return fmt.Errorf("upsert hourly: %w", err)
	}
	return nil
}

// GetByDate 按日期查询全部小时摘要
func (r *PgHourlyRepo) GetByDate(ctx context.Context, userID int, date string) ([]*activity.HourlySummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT hour, summary, main_apps, activity_count, total_duration,
			COALESCE(representative_screenshots, '')
		FROM hourly_summaries
		WHERE user_id = $1 AND date = $2
		ORDER BY hour`,
		userID, date)
	if err != nil {
		return nil, fmt.Errorf("get hourly: %w", err)
	}
	defer rows.Close()

	var items []*activity.HourlySummary
	for rows.Next() {
		s := &activity.HourlySummary{}
		var screenshotsJSON string
		if err := rows.Scan(&s.Hour, &s.Summary, &s.MainApps,
			&s.ActivityCount, &s.TotalDuration, &screenshotsJSON); err != nil {
			return nil, err
		}
		// 简单解析 JSON 数组
		s.RepresentativeScreenshots = parseJSONStringArray(screenshotsJSON)
		items = append(items, s)
	}
	return items, nil
}

// parseJSONStringArray 简单解析 JSON 字符串数组
func parseJSONStringArray(raw string) []string {
	if raw == "" || raw == "[]" {
		return []string{}
	}
	// 简单处理：遍历提取双引号之间的内容
	var result []string
	inQuote := false
	start := 0
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			if inQuote {
				result = append(result, raw[start:i])
				inQuote = false
			} else {
				start = i + 1
				inQuote = true
			}
		}
	}
	return result
}

package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"daylens-server/internal/domain/activity"
)

// PgCategoryRepo 分类规则 PostgreSQL 仓储
type PgCategoryRepo struct {
	pool *pgxpool.Pool
}

// NewPgCategoryRepo 创建分类规则仓储
func NewPgCategoryRepo(pool *pgxpool.Pool) *PgCategoryRepo {
	return &PgCategoryRepo{pool: pool}
}

// Upsert 插入或更新分类规则
func (r *PgCategoryRepo) Upsert(ctx context.Context, userID int, appName, category string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO category_rules (user_id, app_name, category)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, app_name) DO UPDATE SET category = EXCLUDED.category`,
		userID, appName, category)
	if err != nil {
		return fmt.Errorf("upsert category rule: %w", err)
	}
	return nil
}

// GetAll 获取用户全部自定义分类规则
func (r *PgCategoryRepo) GetAll(ctx context.Context, userID int) ([]*activity.CategoryRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT app_name, category
		FROM category_rules
		WHERE user_id = $1
		ORDER BY app_name`, userID)
	if err != nil {
		return nil, fmt.Errorf("get category rules: %w", err)
	}
	defer rows.Close()

	var items []*activity.CategoryRule
	for rows.Next() {
		rule := &activity.CategoryRule{}
		if err := rows.Scan(&rule.AppName, &rule.Category); err != nil {
			return nil, err
		}
		items = append(items, rule)
	}
	return items, nil
}

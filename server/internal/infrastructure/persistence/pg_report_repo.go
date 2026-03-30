package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"daylens-server/internal/domain/report"
)

// PgReportRepo 日报 PostgreSQL 仓储
type PgReportRepo struct {
	pool *pgxpool.Pool
}

// NewPgReportRepo 创建日报仓储
func NewPgReportRepo(pool *pgxpool.Pool) *PgReportRepo {
	return &PgReportRepo{pool: pool}
}

// Upsert 插入或更新日报
func (r *PgReportRepo) Upsert(ctx context.Context, userID int, rpt *report.DailyReport) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO daily_reports (user_id, date, content, ai_mode, model_name, used_ai)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, date) DO UPDATE SET
			content = EXCLUDED.content,
			ai_mode = EXCLUDED.ai_mode,
			model_name = EXCLUDED.model_name,
			used_ai = EXCLUDED.used_ai,
			updated_at = NOW()`,
		userID, rpt.Date, rpt.Content, rpt.AIMode, rpt.ModelName, rpt.UsedAI)
	if err != nil {
		return fmt.Errorf("upsert report: %w", err)
	}
	return nil
}

// GetByDate 按日期查询日报
func (r *PgReportRepo) GetByDate(ctx context.Context, userID int, date string) (*report.DailyReport, error) {
	rpt := &report.DailyReport{}
	err := r.pool.QueryRow(ctx, `
		SELECT date, content, ai_mode, COALESCE(model_name, ''), used_ai,
			EXTRACT(EPOCH FROM created_at)::BIGINT,
			EXTRACT(EPOCH FROM updated_at)::BIGINT
		FROM daily_reports
		WHERE user_id = $1 AND date = $2`,
		userID, date,
	).Scan(&rpt.Date, &rpt.Content, &rpt.AIMode, &rpt.ModelName, &rpt.UsedAI,
		&rpt.CreatedAt, &rpt.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // 日报不存在返回 nil（非错误）
	}
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}
	return rpt, nil
}

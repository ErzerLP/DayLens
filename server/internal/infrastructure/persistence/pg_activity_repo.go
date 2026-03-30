// Package persistence PostgreSQL 数据持久化适配器。
package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"daylens-server/internal/domain/activity"
	"daylens-server/internal/infrastructure/crypto"
	"daylens-server/internal/shared"
)

// PgActivityRepo 活动 PostgreSQL 仓储
type PgActivityRepo struct {
	pool   *pgxpool.Pool
	cipher crypto.FieldCipher
}

// NewPgActivityRepo 创建活动仓储
func NewPgActivityRepo(pool *pgxpool.Pool, cipher crypto.FieldCipher) *PgActivityRepo {
	if cipher == nil {
		cipher = crypto.NopCipher{}
	}
	return &PgActivityRepo{pool: pool, cipher: cipher}
}

// encryptActivity 加密活动的敏感字段（window_title, ocr_text, browser_url）
func (r *PgActivityRepo) encryptActivity(a *activity.Activity) (windowTitle string, ocrText *string, browserURL *string, err error) {
	windowTitle, err = r.cipher.Encrypt(a.WindowTitle)
	if err != nil {
		return "", nil, nil, fmt.Errorf("encrypt window_title: %w", err)
	}
	ocrText, err = r.cipher.EncryptPtr(a.OcrText)
	if err != nil {
		return "", nil, nil, fmt.Errorf("encrypt ocr_text: %w", err)
	}
	browserURL, err = r.cipher.EncryptPtr(a.BrowserURL)
	if err != nil {
		return "", nil, nil, fmt.Errorf("encrypt browser_url: %w", err)
	}
	return
}

// decryptActivity 解密活动的敏感字段
func (r *PgActivityRepo) decryptActivity(a *activity.Activity) error {
	var err error
	a.WindowTitle, err = r.cipher.Decrypt(a.WindowTitle)
	if err != nil {
		return fmt.Errorf("decrypt window_title: %w", err)
	}
	a.OcrText, err = r.cipher.DecryptPtr(a.OcrText)
	if err != nil {
		return fmt.Errorf("decrypt ocr_text: %w", err)
	}
	a.BrowserURL, err = r.cipher.DecryptPtr(a.BrowserURL)
	if err != nil {
		return fmt.Errorf("decrypt browser_url: %w", err)
	}
	return nil
}

// Insert 插入单条活动，幂等（ON CONFLICT DO NOTHING）
func (r *PgActivityRepo) Insert(ctx context.Context, a *activity.Activity) (int64, error) {
	encTitle, encOcr, encURL, err := r.encryptActivity(a)
	if err != nil {
		return 0, fmt.Errorf("insert activity: %w", err)
	}

	var id int64
	err = r.pool.QueryRow(ctx, `
		INSERT INTO activities (user_id, client_id, client_ts, timestamp, app_name, window_title,
			screenshot_key, ocr_text, category, semantic_category, semantic_confidence,
			duration, browser_url, executable_path, extra_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, '{}')
		ON CONFLICT (user_id, client_id, client_ts) DO NOTHING
		RETURNING id`,
		1, a.ClientID, a.ClientTs, a.Timestamp, a.AppName, encTitle,
		ptrStr(a.ScreenshotKey), encOcr, a.Category, a.SemanticCategory,
		a.SemanticConfidence, a.Duration, encURL, a.ExecutablePath,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil // 幂等命中
	}
	if err != nil {
		return 0, fmt.Errorf("insert activity: %w", err)
	}
	return id, nil
}

// InsertBatch 批量插入活动
func (r *PgActivityRepo) InsertBatch(ctx context.Context, list []*activity.Activity) (int, error) {
	inserted := 0
	batch := &pgx.Batch{}

	for _, a := range list {
		encTitle, encOcr, encURL, err := r.encryptActivity(a)
		if err != nil {
			return 0, fmt.Errorf("batch encrypt: %w", err)
		}
		batch.Queue(`
			INSERT INTO activities (user_id, client_id, client_ts, timestamp, app_name, window_title,
				screenshot_key, ocr_text, category, semantic_category, semantic_confidence,
				duration, browser_url, executable_path, extra_json)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, '{}')
			ON CONFLICT (user_id, client_id, client_ts) DO NOTHING`,
			1, a.ClientID, a.ClientTs, a.Timestamp, a.AppName, encTitle,
			ptrStr(a.ScreenshotKey), encOcr, a.Category, a.SemanticCategory,
			a.SemanticConfidence, a.Duration, encURL, a.ExecutablePath,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range list {
		ct, err := br.Exec()
		if err != nil {
			return inserted, fmt.Errorf("batch insert: %w", err)
		}
		if ct.RowsAffected() > 0 {
			inserted++
		}
	}
	return inserted, nil
}

// QueryByDate 按日期分页查询，支持 app/category 筛选
func (r *PgActivityRepo) QueryByDate(ctx context.Context, userID int, date, app, category string, limit, offset int) ([]*activity.Activity, int, error) {
	// 计算日期范围的时间戳
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, 0, fmt.Errorf("parse date: %w", err)
	}
	start := t.Unix()
	end := t.Add(24 * time.Hour).Unix()

	// 构建动态查询
	where := "user_id = $1 AND timestamp >= $2 AND timestamp < $3"
	args := []interface{}{userID, start, end}
	argIdx := 4

	if app != "" {
		where += fmt.Sprintf(" AND app_name = $%d", argIdx)
		args = append(args, app)
		argIdx++
	}
	if category != "" {
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}

	// 查总数
	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM activities WHERE %s", where)
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count activities: %w", err)
	}

	// 分页查询
	querySQL := fmt.Sprintf(`
		SELECT id, client_id, client_ts, timestamp, app_name, window_title,
			screenshot_key, ocr_text, category, semantic_category, semantic_confidence,
			duration, browser_url, executable_path
		FROM activities WHERE %s
		ORDER BY timestamp DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query activities: %w", err)
	}
	defer rows.Close()

	items := make([]*activity.Activity, 0)
	for rows.Next() {
		a := &activity.Activity{}
		if err := rows.Scan(
			&a.ID, &a.ClientID, &a.ClientTs, &a.Timestamp, &a.AppName, &a.WindowTitle,
			&a.ScreenshotKey, &a.OcrText, &a.Category, &a.SemanticCategory,
			&a.SemanticConfidence, &a.Duration, &a.BrowserURL, &a.ExecutablePath,
		); err != nil {
			return nil, 0, fmt.Errorf("scan activity: %w", err)
		}
		// 解密敏感字段
		if err := r.decryptActivity(a); err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, nil
}

// GetByID 按 ID 查询单条活动
func (r *PgActivityRepo) GetByID(ctx context.Context, id int64) (*activity.Activity, error) {
	a := &activity.Activity{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, client_id, client_ts, timestamp, app_name, window_title,
			screenshot_key, ocr_text, category, semantic_category, semantic_confidence,
			duration, browser_url, executable_path
		FROM activities WHERE id = $1`, id,
	).Scan(
		&a.ID, &a.ClientID, &a.ClientTs, &a.Timestamp, &a.AppName, &a.WindowTitle,
		&a.ScreenshotKey, &a.OcrText, &a.Category, &a.SemanticCategory,
		&a.SemanticConfidence, &a.Duration, &a.BrowserURL, &a.ExecutablePath,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, shared.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get activity %d: %w", id, err)
	}
	// 解密敏感字段
	if err := r.decryptActivity(a); err != nil {
		return nil, err
	}
	return a, nil
}

// GetDailyStats 获取指定日期的聚合统计
func (r *PgActivityRepo) GetDailyStats(ctx context.Context, userID int, date string) (*activity.DailyStats, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("parse date: %w", err)
	}
	start := t.Unix()
	end := t.Add(24 * time.Hour).Unix()

	stats := &activity.DailyStats{Date: date}

	// 总时长 + 截图数 + 活跃小时 + 工作时段时长
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(duration), 0),
			   COUNT(CASE WHEN screenshot_key != '' THEN 1 END),
			   COUNT(DISTINCT EXTRACT(HOUR FROM TO_TIMESTAMP(timestamp))),
			   COALESCE(SUM(CASE WHEN EXTRACT(HOUR FROM TO_TIMESTAMP(timestamp)) BETWEEN 9 AND 17 THEN duration ELSE 0 END), 0)
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3`,
		userID, start, end,
	).Scan(&stats.TotalDuration, &stats.ScreenshotCount, &stats.ActiveHours, &stats.WorkTimeDuration)
	if err != nil {
		return nil, fmt.Errorf("stats aggregation: %w", err)
	}

	// 应用使用分布
	stats.AppUsage, err = r.queryAppUsage(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	// 分类使用分布
	stats.CategoryUsage, err = r.queryCategoryUsage(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	// 域名使用分布（加密时无法在 SQL 层解析——降级为空）
	stats.DomainUsage, err = r.queryDomainUsage(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	// 最常用窗口标题（加密时在应用层解密聚合）
	stats.TopWindowTitles, err = r.queryTopTitles(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// queryAppUsage 查询应用使用统计
func (r *PgActivityRepo) queryAppUsage(ctx context.Context, userID int, start, end int64) ([]activity.AppUsage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT app_name, COALESCE(SUM(duration), 0), COUNT(*)
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3
		GROUP BY app_name ORDER BY SUM(duration) DESC LIMIT 20`,
		userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("app usage: %w", err)
	}
	defer rows.Close()

	var items []activity.AppUsage
	for rows.Next() {
		var u activity.AppUsage
		if err := rows.Scan(&u.AppName, &u.Duration, &u.Count); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, nil
}

// queryCategoryUsage 查询分类使用统计
func (r *PgActivityRepo) queryCategoryUsage(ctx context.Context, userID int, start, end int64) ([]activity.CategoryUsage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT category, COALESCE(SUM(duration), 0)
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3
		GROUP BY category ORDER BY SUM(duration) DESC`,
		userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("category usage: %w", err)
	}
	defer rows.Close()

	var items []activity.CategoryUsage
	for rows.Next() {
		var u activity.CategoryUsage
		if err := rows.Scan(&u.Category, &u.Duration); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, nil
}

// queryDomainUsage 查询域名使用统计
func (r *PgActivityRepo) queryDomainUsage(ctx context.Context, userID int, start, end int64) ([]activity.DomainUsage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			CASE
				WHEN browser_url IS NOT NULL AND browser_url != ''
				THEN SPLIT_PART(SPLIT_PART(browser_url, '://', 2), '/', 1)
				ELSE ''
			END AS domain,
			COALESCE(SUM(duration), 0)
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3
			AND browser_url IS NOT NULL AND browser_url != ''
		GROUP BY domain ORDER BY SUM(duration) DESC LIMIT 10`,
		userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("domain usage: %w", err)
	}
	defer rows.Close()

	var items []activity.DomainUsage
	for rows.Next() {
		var u activity.DomainUsage
		if err := rows.Scan(&u.Domain, &u.Duration); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, nil
}

// queryTopTitles 查询最常用窗口标题
func (r *PgActivityRepo) queryTopTitles(ctx context.Context, userID int, start, end int64) ([]activity.WindowTitleUsage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT window_title, COALESCE(SUM(duration), 0), app_name
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3
		GROUP BY window_title, app_name ORDER BY SUM(duration) DESC LIMIT 10`,
		userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("top titles: %w", err)
	}
	defer rows.Close()

	var items []activity.WindowTitleUsage
	for rows.Next() {
		var u activity.WindowTitleUsage
		if err := rows.Scan(&u.Title, &u.Duration, &u.AppName); err != nil {
			return nil, err
		}
		// 解密标题
		u.Title, _ = r.cipher.Decrypt(u.Title)
		items = append(items, u)
	}
	return items, nil
}

// Search 全文搜索活动（LIKE + 权重评分）
func (r *PgActivityRepo) Search(ctx context.Context, userID int, query string, limit int) ([]*activity.SearchResultItem, int, error) {
	pattern := "%" + query + "%"

	rows, err := r.pool.Query(ctx, `
		SELECT id, timestamp, app_name,
			CASE
				WHEN window_title ILIKE $2 THEN window_title
				WHEN ocr_text ILIKE $2 THEN LEFT(ocr_text, 200)
				WHEN browser_url ILIKE $2 THEN browser_url
				ELSE ''
			END AS excerpt,
			CASE
				WHEN window_title ILIKE $2 THEN 'windowTitle'
				WHEN ocr_text ILIKE $2 THEN 'ocrText'
				WHEN browser_url ILIKE $2 THEN 'browserUrl'
				ELSE 'unknown'
			END AS match_field,
			CASE
				WHEN window_title ILIKE $2 THEN 100
				WHEN ocr_text ILIKE $2 THEN 80
				WHEN browser_url ILIKE $2 THEN 60
				ELSE 0
			END AS relevance_score
		FROM activities
		WHERE user_id = $1
			AND (window_title ILIKE $2 OR ocr_text ILIKE $2 OR browser_url ILIKE $2)
		ORDER BY relevance_score DESC, timestamp DESC
		LIMIT $3`,
		userID, pattern, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	var items []*activity.SearchResultItem
	for rows.Next() {
		item := &activity.SearchResultItem{}
		if err := rows.Scan(&item.ActivityID, &item.Timestamp, &item.AppName,
			&item.Excerpt, &item.MatchField, &item.RelevanceScore); err != nil {
			return nil, 0, err
		}
		// 解密搜索结果摘要
		item.Excerpt, _ = r.cipher.Decrypt(item.Excerpt)
		items = append(items, item)
	}

	// 查总匹配数
	var total int
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM activities
		WHERE user_id = $1
			AND (window_title ILIKE $2 OR ocr_text ILIKE $2 OR browser_url ILIKE $2)`,
		userID, pattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// DeleteBefore 删除指定时间之前的活动
func (r *PgActivityRepo) DeleteBefore(ctx context.Context, userID int, before time.Time) (int64, error) {
	ct, err := r.pool.Exec(ctx, `
		DELETE FROM activities WHERE user_id = $1 AND timestamp < $2`,
		userID, before.Unix())
	if err != nil {
		return 0, fmt.Errorf("delete before: %w", err)
	}
	return ct.RowsAffected(), nil
}

// GetRecentApps 获取最近 N 天使用过的应用列表
func (r *PgActivityRepo) GetRecentApps(ctx context.Context, userID int, days int) ([]string, error) {
	since := time.Now().AddDate(0, 0, -days).Unix()
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT app_name FROM activities
		WHERE user_id = $1 AND timestamp >= $2
		ORDER BY app_name`, userID, since)
	if err != nil {
		return nil, fmt.Errorf("recent apps: %w", err)
	}
	defer rows.Close()

	var apps []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		apps = append(apps, name)
	}
	return apps, nil
}

// GetAppCategories 获取应用分类概览
func (r *PgActivityRepo) GetAppCategories(ctx context.Context, userID int, from, to string) ([]*activity.AppCategoryInfo, error) {
	fromT, _ := time.Parse("2006-01-02", from)
	toT, _ := time.Parse("2006-01-02", to)
	toT = toT.Add(24 * time.Hour)

	rows, err := r.pool.Query(ctx, `
		SELECT app_name, category,
			COALESCE(SUM(duration), 0) AS total_duration,
			MAX(timestamp) AS last_seen
		FROM activities
		WHERE user_id = $1 AND timestamp >= $2 AND timestamp < $3
		GROUP BY app_name, category
		ORDER BY total_duration DESC`,
		userID, fromT.Unix(), toT.Unix())
	if err != nil {
		return nil, fmt.Errorf("app categories: %w", err)
	}
	defer rows.Close()

	var items []*activity.AppCategoryInfo
	for rows.Next() {
		info := &activity.AppCategoryInfo{}
		if err := rows.Scan(&info.AppName, &info.Category, &info.TotalDuration, &info.LastSeen); err != nil {
			return nil, err
		}
		items = append(items, info)
	}
	return items, nil
}

// Reclassify 更新指定应用的分类
func (r *PgActivityRepo) Reclassify(ctx context.Context, userID int, appName, newCategory string) (int64, error) {
	ct, err := r.pool.Exec(ctx, `
		UPDATE activities SET category = $3
		WHERE user_id = $1 AND app_name = $2`,
		userID, appName, newCategory)
	if err != nil {
		return 0, fmt.Errorf("reclassify: %w", err)
	}
	return ct.RowsAffected(), nil
}

// ptrStr 安全取 *string 值
func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

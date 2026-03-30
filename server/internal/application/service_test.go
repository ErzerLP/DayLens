package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/report"
)

// ===== Mock 实现 =====

// mockActivityRepo 模拟活动仓储
type mockActivityRepo struct {
	insertFn       func(ctx context.Context, a *activity.Activity) (int64, error)
	insertBatchFn  func(ctx context.Context, list []*activity.Activity) (int, error)
	queryByDateFn  func(ctx context.Context, uid int, date, app, category string, limit, offset int) ([]*activity.Activity, int, error)
	getByIDFn      func(ctx context.Context, id int64) (*activity.Activity, error)
	getDailyStatsFn func(ctx context.Context, uid int, date string) (*activity.DailyStats, error)
	searchFn       func(ctx context.Context, uid int, query string, limit int) ([]*activity.SearchResultItem, int, error)
	deleteBefore   func(ctx context.Context, uid int, before time.Time) (int64, error)
	recentAppsFn   func(ctx context.Context, uid int, days int) ([]string, error)
	appCatFn       func(ctx context.Context, uid int, from, to string) ([]*activity.AppCategoryInfo, error)
	reclassifyFn   func(ctx context.Context, uid int, appName, newCategory string) (int64, error)
}

func (m *mockActivityRepo) Insert(ctx context.Context, a *activity.Activity) (int64, error) {
	if m.insertFn != nil { return m.insertFn(ctx, a) }
	return 1, nil
}
func (m *mockActivityRepo) InsertBatch(ctx context.Context, list []*activity.Activity) (int, error) {
	if m.insertBatchFn != nil { return m.insertBatchFn(ctx, list) }
	return len(list), nil
}
func (m *mockActivityRepo) QueryByDate(ctx context.Context, uid int, date, app, category string, limit, offset int) ([]*activity.Activity, int, error) {
	if m.queryByDateFn != nil { return m.queryByDateFn(ctx, uid, date, app, category, limit, offset) }
	return nil, 0, nil
}
func (m *mockActivityRepo) GetByID(ctx context.Context, id int64) (*activity.Activity, error) {
	if m.getByIDFn != nil { return m.getByIDFn(ctx, id) }
	return nil, nil
}
func (m *mockActivityRepo) GetDailyStats(ctx context.Context, uid int, date string) (*activity.DailyStats, error) {
	if m.getDailyStatsFn != nil { return m.getDailyStatsFn(ctx, uid, date) }
	return &activity.DailyStats{Date: date, TotalDuration: 3600}, nil
}
func (m *mockActivityRepo) Search(ctx context.Context, uid int, query string, limit int) ([]*activity.SearchResultItem, int, error) {
	if m.searchFn != nil { return m.searchFn(ctx, uid, query, limit) }
	return nil, 0, nil
}
func (m *mockActivityRepo) DeleteBefore(ctx context.Context, uid int, before time.Time) (int64, error) {
	if m.deleteBefore != nil { return m.deleteBefore(ctx, uid, before) }
	return 0, nil
}
func (m *mockActivityRepo) GetRecentApps(ctx context.Context, uid int, days int) ([]string, error) {
	if m.recentAppsFn != nil { return m.recentAppsFn(ctx, uid, days) }
	return nil, nil
}
func (m *mockActivityRepo) GetAppCategories(ctx context.Context, uid int, from, to string) ([]*activity.AppCategoryInfo, error) {
	if m.appCatFn != nil { return m.appCatFn(ctx, uid, from, to) }
	return nil, nil
}
func (m *mockActivityRepo) Reclassify(ctx context.Context, uid int, appName, newCategory string) (int64, error) {
	if m.reclassifyFn != nil { return m.reclassifyFn(ctx, uid, appName, newCategory) }
	return 0, nil
}

// mockReportRepo 模拟日报仓储
type mockReportRepo struct {
	upsertFn    func(ctx context.Context, uid int, r *report.DailyReport) error
	getByDateFn func(ctx context.Context, uid int, date string) (*report.DailyReport, error)
}

func (m *mockReportRepo) Upsert(ctx context.Context, uid int, r *report.DailyReport) error {
	if m.upsertFn != nil { return m.upsertFn(ctx, uid, r) }
	return nil
}
func (m *mockReportRepo) GetByDate(ctx context.Context, uid int, date string) (*report.DailyReport, error) {
	if m.getByDateFn != nil { return m.getByDateFn(ctx, uid, date) }
	return nil, nil
}

// mockEventBus 模拟事件总线
type mockEventBus struct {
	published []string
}

func (m *mockEventBus) Publish(event string, _ interface{}) {
	m.published = append(m.published, event)
}

// mockAIProvider 模拟 AI 提供商
type mockAIProvider struct {
	available bool
	reply     string
	err       error
}

func (m *mockAIProvider) Name() string { return "mock" }
func (m *mockAIProvider) Chat(_ context.Context, _ []port.Message) (string, error) {
	return m.reply, m.err
}
func (m *mockAIProvider) IsAvailable(_ context.Context) bool { return m.available }

// ===== ActivityService 测试 =====

// 正常上报活动应返回新 ID 并推送事件
func TestActivityService_IngestActivity_Success(t *testing.T) {
	bus := &mockEventBus{}
	svc := NewActivityService(&mockActivityRepo{
		insertFn: func(_ context.Context, _ *activity.Activity) (int64, error) {
			return 42, nil
		},
	}, bus)

	id, err := svc.IngestActivity(context.Background(), &activity.Activity{AppName: "Code"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("got id %d, want 42", id)
	}
	if len(bus.published) != 1 || bus.published[0] != "activity_received" {
		t.Errorf("expected activity_received event, got %v", bus.published)
	}
}

// 缺少 appName 应返回错误
func TestActivityService_IngestActivity_MissingAppName(t *testing.T) {
	svc := NewActivityService(&mockActivityRepo{}, &mockEventBus{})
	_, err := svc.IngestActivity(context.Background(), &activity.Activity{})
	if err == nil {
		t.Fatal("expected error for missing appName")
	}
	if !errors.Is(err, ErrFieldMissing) {
		t.Errorf("got %v, want ErrFieldMissing", err)
	}
}

// 幂等上报（id=0）不应推送事件
func TestActivityService_IngestActivity_Deduplicated(t *testing.T) {
	bus := &mockEventBus{}
	svc := NewActivityService(&mockActivityRepo{
		insertFn: func(_ context.Context, _ *activity.Activity) (int64, error) {
			return 0, nil // 幂等命中
		},
	}, bus)

	id, err := svc.IngestActivity(context.Background(), &activity.Activity{AppName: "Code"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 0 {
		t.Errorf("got id %d, want 0", id)
	}
	if len(bus.published) != 0 {
		t.Errorf("should not publish on dedup, got %v", bus.published)
	}
}

// 批量上报应返回正确的插入和去重计数
func TestActivityService_IngestBatch_Success(t *testing.T) {
	svc := NewActivityService(&mockActivityRepo{
		insertBatchFn: func(_ context.Context, list []*activity.Activity) (int, error) {
			return 3, nil // 5 条中 3 条新插入
		},
	}, &mockEventBus{})

	list := make([]*activity.Activity, 5)
	for i := range list {
		list[i] = &activity.Activity{AppName: "Code"}
	}

	result, err := svc.IngestBatch(context.Background(), list)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 5 || result.Inserted != 3 || result.Deduplicated != 2 {
		t.Errorf("got %+v, want total=5 inserted=3 dedup=2", result)
	}
}

// 批量超过 100 条应返回错误
func TestActivityService_IngestBatch_OverLimit(t *testing.T) {
	svc := NewActivityService(&mockActivityRepo{}, &mockEventBus{})
	list := make([]*activity.Activity, 101)
	_, err := svc.IngestBatch(context.Background(), list)
	if err == nil {
		t.Fatal("expected error for over limit")
	}
}

// ===== ReportService 测试 =====

// AI 不可用时应降级生成模板日报
func TestReportService_Generate_FallbackToTemplate(t *testing.T) {
	svc := NewReportService(
		&mockReportRepo{},
		&mockActivityRepo{},
		&mockAIProvider{available: false},
		&mockEventBus{},
	)

	r, err := svc.Generate(context.Background(), 1, "2026-03-29", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.UsedAI {
		t.Error("expected usedAI=false for template fallback")
	}
	if r.AIMode != "template" {
		t.Errorf("got aiMode=%s, want template", r.AIMode)
	}
}

// 非强制生成时已有日报应直接返回
func TestReportService_Generate_ExistingReport(t *testing.T) {
	existing := &report.DailyReport{Date: "2026-03-29", Content: "已有日报"}
	svc := NewReportService(
		&mockReportRepo{
			getByDateFn: func(_ context.Context, _ int, _ string) (*report.DailyReport, error) {
				return existing, nil
			},
		},
		&mockActivityRepo{},
		&mockAIProvider{available: false},
		&mockEventBus{},
	)

	r, err := svc.Generate(context.Background(), 1, "2026-03-29", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Content != "已有日报" {
		t.Errorf("should return existing report")
	}
}

package application

import (
	"context"
	"fmt"
	"log/slog"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/report"
)

// ReportService 日报生成与查询用例编排
type ReportService struct {
	reportRepo   port.ReportRepository
	activityRepo port.ActivityRepository
	aiProvider   port.AIProvider
	eventBus     port.EventBus
}

// NewReportService 创建日报服务
func NewReportService(
	reportRepo port.ReportRepository,
	activityRepo port.ActivityRepository,
	aiProvider port.AIProvider,
	eventBus port.EventBus,
) *ReportService {
	return &ReportService{
		reportRepo:   reportRepo,
		activityRepo: activityRepo,
		aiProvider:   aiProvider,
		eventBus:     eventBus,
	}
}

// GetByDate 查询指定日期的日报
func (s *ReportService) GetByDate(ctx context.Context, userID int, date string) (*report.DailyReport, error) {
	r, err := s.reportRepo.GetByDate(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("get report: %w", err)
	}
	return r, nil
}

// Generate 生成日报，AI 不可用时降级为模板
func (s *ReportService) Generate(ctx context.Context, userID int, date string, forceRegenerate bool) (*report.DailyReport, error) {
	// 非强制时检查是否已存在
	if !forceRegenerate {
		existing, err := s.reportRepo.GetByDate(ctx, userID, date)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	// 获取当天统计数据
	stats, err := s.activityRepo.GetDailyStats(ctx, userID, date)
	if err != nil {
		return nil, fmt.Errorf("generate report: %w", err)
	}

	// 尝试 AI 生成
	var r *report.DailyReport
	if s.aiProvider.IsAvailable(ctx) {
		r, err = s.generateWithAI(ctx, date, stats)
		if err != nil {
			slog.Warn("AI 生成日报失败，降级为模板", "error", err, "date", date)
			r = s.generateTemplate(date, stats)
		}
	} else {
		r = s.generateTemplate(date, stats)
	}

	// 持久化
	if err := s.reportRepo.Upsert(ctx, userID, r); err != nil {
		return nil, fmt.Errorf("save report: %w", err)
	}

	s.eventBus.Publish("report_generated", map[string]interface{}{"date": date})
	return r, nil
}

// ExportMarkdown 导出日报为 Markdown 文本
func (s *ReportService) ExportMarkdown(ctx context.Context, userID int, date string) (string, error) {
	r, err := s.reportRepo.GetByDate(ctx, userID, date)
	if err != nil {
		return "", fmt.Errorf("export report: %w", err)
	}
	if r == nil {
		return "", ErrNotFound
	}
	return r.Content, nil
}

// generateWithAI 调用 AI 生成日报
func (s *ReportService) generateWithAI(ctx context.Context, date string, stats *activity.DailyStats) (*report.DailyReport, error) {
	prompt := buildReportPrompt(date, stats)
	messages := []port.Message{
		{Role: "system", Content: "你是一个工作总结助手，负责根据用户的电脑使用数据生成简洁的每日工作总结。使用 Markdown 格式。"},
		{Role: "user", Content: prompt},
	}

	content, err := s.aiProvider.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	return &report.DailyReport{
		Date:      date,
		Content:   content,
		AIMode:    "summary",
		ModelName: s.aiProvider.Name(),
		UsedAI:    true,
	}, nil
}

// generateTemplate 使用模板生成日报（AI 降级）
func (s *ReportService) generateTemplate(date string, stats *activity.DailyStats) *report.DailyReport {
	content := fmt.Sprintf("# %s 工作日报\n\n- 总活跃时长: %d 分钟\n- 活跃小时数: %d\n- 截图数: %d\n",
		date, stats.TotalDuration/60, stats.ActiveHours, stats.ScreenshotCount)

	if len(stats.AppUsage) > 0 {
		content += "\n## 应用使用\n\n"
		for _, app := range stats.AppUsage {
			content += fmt.Sprintf("- **%s**: %d 分钟\n", app.AppName, app.Duration/60)
		}
	}

	return &report.DailyReport{
		Date:    date,
		Content: content,
		AIMode:  "template",
		UsedAI:  false,
	}
}

// buildReportPrompt 构建日报生成提示词
func buildReportPrompt(date string, stats *activity.DailyStats) string {
	prompt := fmt.Sprintf("请根据以下 %s 的工作数据生成日报：\n\n", date)
	prompt += fmt.Sprintf("总活跃时长: %d 分钟\n活跃小时数: %d\n\n", stats.TotalDuration/60, stats.ActiveHours)

	if len(stats.AppUsage) > 0 {
		prompt += "应用使用情况:\n"
		for _, app := range stats.AppUsage {
			prompt += fmt.Sprintf("- %s: %d 分钟 (%d 次)\n", app.AppName, app.Duration/60, app.Count)
		}
	}

	if len(stats.CategoryUsage) > 0 {
		prompt += "\n分类分布:\n"
		for _, cat := range stats.CategoryUsage {
			prompt += fmt.Sprintf("- %s: %d 分钟\n", cat.Category, cat.Duration/60)
		}
	}

	return prompt
}

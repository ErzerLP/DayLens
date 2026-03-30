// Package scheduler 定时任务调度。
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"daylens-server/config"
	"daylens-server/internal/application"
	"daylens-server/internal/application/port"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron       *cron.Cron
	reportSvc  *application.ReportService
	storageSvc *application.StorageService
	hourlyRepo port.HourlyRepository
	activityRepo port.ActivityRepository
	cfg        *config.Config
}

// New 创建调度器
func New(
	cfg *config.Config,
	reportSvc *application.ReportService,
	storageSvc *application.StorageService,
	hourlyRepo port.HourlyRepository,
	activityRepo port.ActivityRepository,
) *Scheduler {
	return &Scheduler{
		cron:         cron.New(),
		reportSvc:    reportSvc,
		storageSvc:   storageSvc,
		hourlyRepo:   hourlyRepo,
		activityRepo: activityRepo,
		cfg:          cfg,
	}
}

// Start 注册并启动定时任务
func (s *Scheduler) Start() {
	// 日报定时生成（工作结束时间）
	reportExpr := fmt.Sprintf("%d %d * * *", s.cfg.Schedule.EndMinute, s.cfg.Schedule.EndHour)
	_, _ = s.cron.AddFunc(reportExpr, s.generateDailyReport)
	slog.Info("定时日报已注册", "cron", reportExpr)

	// 小时摘要更新（每小时第 5 分钟）
	_, _ = s.cron.AddFunc("5 * * * *", s.updateHourlySummary)
	slog.Info("小时摘要任务已注册", "cron", "5 * * * *")

	// 数据清理（凌晨 3 点）
	_, _ = s.cron.AddFunc("0 3 * * *", s.cleanupOldData)
	slog.Info("数据清理任务已注册", "cron", "0 3 * * *")

	s.cron.Start()
	slog.Info("定时任务调度器已启动")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("定时任务调度器已停止")
}

// generateDailyReport 定时生成当日日报
func (s *Scheduler) generateDailyReport() {
	date := time.Now().Format("2006-01-02")
	slog.Info("开始定时生成日报", "date", date)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, err := s.reportSvc.Generate(ctx, 1, date, false)
	if err != nil {
		slog.Error("定时日报生成失败", "date", date, "error", err)
		return
	}
	slog.Info("定时日报生成完成", "date", date)
}

// updateHourlySummary 更新当前小时的摘要
func (s *Scheduler) updateHourlySummary() {
	now := time.Now()
	date := now.Format("2006-01-02")
	hour := now.Hour() - 1 // 统计上一小时
	if hour < 0 {
		return
	}

	slog.Info("开始更新小时摘要", "date", date, "hour", hour)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 查询上一小时的活动
	hourStart := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	hourEnd := hourStart.Add(time.Hour)

	items, _, err := s.activityRepo.QueryByDate(ctx, 1, date, "", "", 2000, 0)
	if err != nil {
		slog.Error("查询小时活动失败", "error", err)
		return
	}

	// 筛选指定小时内的活动
	var count int64
	var totalDuration int64
	appMap := make(map[string]int64)
	for _, a := range items {
		ts := time.Unix(a.Timestamp, 0)
		if ts.Before(hourStart) || ts.After(hourEnd) {
			continue
		}
		count++
		totalDuration += int64(a.Duration)
		appMap[a.AppName] += int64(a.Duration)
	}

	if count == 0 {
		return // 该小时无活动
	}

	// 构建 mainApps 字符串
	mainApps := buildMainApps(appMap)
	summary := fmt.Sprintf("本小时共 %d 条活动，总时长 %d 分钟", count, totalDuration/60)

	hourSummary := &struct {
		Hour                      int
		Summary                   string
		MainApps                  string
		ActivityCount             int64
		TotalDuration             int64
		RepresentativeScreenshots []string
	}{
		Hour:          hour,
		Summary:       summary,
		MainApps:      mainApps,
		ActivityCount: count,
		TotalDuration: totalDuration,
	}

	// 转为 domain 类型
	import_activity_hourly := toHourlySummaryDomain(hourSummary)
	if err := s.hourlyRepo.Upsert(ctx, 1, import_activity_hourly, date); err != nil {
		slog.Error("更新小时摘要失败", "error", err)
	}
}

// cleanupOldData 清理过期数据
func (s *Scheduler) cleanupOldData() {
	slog.Info("开始清理过期数据", "retentionDays", s.cfg.Storage.RetentionDays)

	beforeDate := time.Now().AddDate(0, 0, -s.cfg.Storage.RetentionDays).Format("2006-01-02")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	result, err := s.storageSvc.DeleteBefore(ctx, 1, beforeDate)
	if err != nil {
		slog.Error("清理过期数据失败", "error", err)
		return
	}
	slog.Info("清理过期数据完成",
		"deletedActivities", result.DeletedActivities,
		"deletedScreenshots", result.DeletedScreenshots)
}

// buildMainApps 从应用使用统计中提取 Top3
func buildMainApps(appMap map[string]int64) string {
	type kv struct {
		key string
		val int64
	}
	var sorted []kv
	for k, v := range appMap {
		sorted = append(sorted, kv{k, v})
	}
	// 简单冒泡排序 Top3
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].val > sorted[i].val {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	var names []string
	for i, item := range sorted {
		if i >= 3 {
			break
		}
		names = append(names, item.key)
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

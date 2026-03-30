package intelligence

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/session"
)

const (
	// deepWorkThreshold 深度工作阈值（秒），会话 >60 分钟且主导应用使用 >70%
	deepWorkThreshold = 60 * 60
	// deepWorkDominanceRatio 主导应用占比阈值
	deepWorkDominanceRatio = 70
)

// GenerateWeeklyReview 生成周报
func GenerateWeeklyReview(activitiesByDay map[string][]*activity.Activity, from, to string) *session.WeeklyReview {
	period := fmt.Sprintf("%s ~ %s", from, to)
	var totalDuration int64
	var allSessions []session.WorkSession
	appDuration := make(map[string]int64)
	var deepSessions []session.DeepWorkSession

	// 按天处理
	dayCount := 0
	current, _ := time.Parse("2006-01-02", from)
	end, _ := time.Parse("2006-01-02", to)
	end = end.Add(24 * time.Hour)

	for current.Before(end) {
		date := current.Format("2006-01-02")
		dayActivities := activitiesByDay[date]
		current = current.Add(24 * time.Hour)

		if len(dayActivities) == 0 {
			continue
		}
		dayCount++

		// 构建当天会话
		sessions := BuildWorkSessions(dayActivities)
		allSessions = append(allSessions, sessions...)

		// 统计总时长
		for _, a := range dayActivities {
			totalDuration += int64(a.Duration)
			appDuration[a.AppName] += int64(a.Duration)
		}

		// 识别深度工作段
		for _, s := range sessions {
			if s.TotalDuration >= deepWorkThreshold {
				// 检查主导应用占比
				dominantDur := dominantAppDuration(s)
				ratio := int(dominantDur * 100 / s.TotalDuration)
				if ratio >= deepWorkDominanceRatio {
					deepSessions = append(deepSessions, session.DeepWorkSession{
						Date:     date,
						Duration: s.TotalDuration,
						Focus:    s.DominantApp,
					})
				}
			}
		}
	}

	// 日均时长
	avgDaily := int64(0)
	if dayCount > 0 {
		avgDaily = totalDuration / int64(dayCount)
	}

	// Top 应用
	topApps := buildTopApps(appDuration, 10)

	// 意图分布
	intentResult := AnalyzeIntents(allSessions)

	// 生成 Markdown 内容
	content := buildWeeklyContent(period, totalDuration, avgDaily, dayCount, topApps, deepSessions, intentResult)

	return &session.WeeklyReview{
		Period:             period,
		TotalWorkDuration:  totalDuration,
		AvgDailyDuration:   avgDaily,
		Content:            content,
		DeepWorkSessions:   deepSessions,
		TopApps:            topApps,
		IntentDistribution: intentResult.Items,
	}
}

// dominantAppDuration 计算会话中主导应用的总时长
func dominantAppDuration(s session.WorkSession) int64 {
	appDur := make(map[string]int64)
	for _, a := range s.Activities {
		appDur[a.AppName] += a.Duration
	}
	var maxDur int64
	for _, dur := range appDur {
		if dur > maxDur {
			maxDur = dur
		}
	}
	return maxDur
}

// buildTopApps 构建 Top N 应用列表
func buildTopApps(appDuration map[string]int64, limit int) []activity.AppUsage {
	type kv struct {
		app string
		dur int64
	}
	var sorted []kv
	for k, v := range appDuration {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].dur > sorted[j].dur })

	var result []activity.AppUsage
	for i, item := range sorted {
		if i >= limit {
			break
		}
		result = append(result, activity.AppUsage{
			AppName:  item.app,
			Duration: item.dur,
			Count:    0,
		})
	}
	return result
}

// buildWeeklyContent 生成周报 Markdown
func buildWeeklyContent(
	period string,
	total, avgDaily int64,
	dayCount int,
	topApps []activity.AppUsage,
	deepSessions []session.DeepWorkSession,
	intents *session.IntentAnalysisResult,
) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# 周报 %s\n\n", period))
	b.WriteString(fmt.Sprintf("## 概览\n\n"))
	b.WriteString(fmt.Sprintf("- **总工作时长**: %.1f 小时\n", float64(total)/3600))
	b.WriteString(fmt.Sprintf("- **日均工作时长**: %.1f 小时\n", float64(avgDaily)/3600))
	b.WriteString(fmt.Sprintf("- **活跃天数**: %d 天\n", dayCount))
	if intents != nil && intents.DominantIntent != "" {
		b.WriteString(fmt.Sprintf("- **主要意图**: %s\n", intents.DominantIntent))
	}

	// Top 应用
	if len(topApps) > 0 {
		b.WriteString("\n## 常用应用\n\n")
		b.WriteString("| 应用 | 时长 |\n|:--|:--|\n")
		for _, app := range topApps {
			b.WriteString(fmt.Sprintf("| %s | %.1f 小时 |\n", app.AppName, float64(app.Duration)/3600))
		}
	}

	// 深度工作
	if len(deepSessions) > 0 {
		b.WriteString("\n## 深度工作\n\n")
		for _, ds := range deepSessions {
			b.WriteString(fmt.Sprintf("- **%s** %s 专注 %.1f 小时\n",
				ds.Date, ds.Focus, float64(ds.Duration)/3600))
		}
	}

	// 意图分布
	if intents != nil && len(intents.Items) > 0 {
		b.WriteString("\n## 意图分布\n\n")
		b.WriteString("| 意图 | 时长 | 占比 |\n|:--|:--|:--|\n")
		for _, item := range intents.Items {
			b.WriteString(fmt.Sprintf("| %s | %.1f 小时 | %d%% |\n",
				item.Label, float64(item.TotalDuration)/3600, item.Percentage))
		}
	}

	return b.String()
}

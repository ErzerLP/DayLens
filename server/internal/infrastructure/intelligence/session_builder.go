// Package intelligence 工作智能引擎（从 Rust 1:1 移植）。
package intelligence

import (
	"sort"

	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/session"
)

const (
	// sessionGapThreshold 会话间隔阈值（秒），超过则拆分为新会话
	sessionGapThreshold = 15 * 60 // 15 分钟
)

// BuildWorkSessions 将活动列表聚合为工作会话
// 算法：按时间排序，相邻活动间隔 >15 分钟则拆分
func BuildWorkSessions(activities []*activity.Activity) []session.WorkSession {
	if len(activities) == 0 {
		return nil
	}

	// 按时间戳排序
	sorted := make([]*activity.Activity, len(activities))
	copy(sorted, activities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp < sorted[j].Timestamp
	})

	var sessions []session.WorkSession
	var current []session.SessionActivity
	var currentActivities []*activity.Activity
	sessionStart := sorted[0].Timestamp

	for i, a := range sorted {
		sa := session.SessionActivity{
			AppName:  a.AppName,
			Duration: int64(a.Duration),
			Title:    a.WindowTitle,
		}

		if i > 0 {
			prevEnd := sorted[i-1].Timestamp + int64(sorted[i-1].Duration)
			gap := a.Timestamp - prevEnd
			if gap > sessionGapThreshold {
				// 间隔过大，结束当前会话
				s := buildSession(sessionStart, sorted[i-1], current, currentActivities)
				sessions = append(sessions, s)
				current = nil
				currentActivities = nil
				sessionStart = a.Timestamp
			}
		}

		current = append(current, sa)
		currentActivities = append(currentActivities, a)
	}

	// 最后一个会话
	if len(current) > 0 {
		last := sorted[len(sorted)-1]
		s := buildSession(sessionStart, last, current, currentActivities)
		sessions = append(sessions, s)
	}

	return sessions
}

// buildSession 构建单个会话
func buildSession(startTime int64, lastActivity *activity.Activity, activities []session.SessionActivity, raw []*activity.Activity) session.WorkSession {
	endTime := lastActivity.Timestamp + int64(lastActivity.Duration)
	totalDuration := endTime - startTime

	// 找主导应用
	appDuration := make(map[string]int64)
	for _, a := range activities {
		appDuration[a.AppName] += a.Duration
	}
	dominant := findDominantApp(appDuration)

	// 意图分析
	intent := ClassifySession(activities, raw)

	return session.WorkSession{
		StartTime:     startTime,
		EndTime:       endTime,
		TotalDuration: totalDuration,
		Activities:    activities,
		DominantApp:   dominant,
		Intent:        intent,
	}
}

// findDominantApp 查找使用时间最长的应用
func findDominantApp(appDuration map[string]int64) string {
	var maxApp string
	var maxDur int64
	for app, dur := range appDuration {
		if dur > maxDur {
			maxDur = dur
			maxApp = app
		}
	}
	return maxApp
}

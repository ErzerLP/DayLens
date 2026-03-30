package application

import (
	"context"
	"fmt"
	"time"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/session"
	"daylens-server/internal/infrastructure/intelligence"
)

// SessionService 工作会话与智能分析用例编排
type SessionService struct {
	activityRepo port.ActivityRepository
}

// NewSessionService 创建会话服务
func NewSessionService(activityRepo port.ActivityRepository) *SessionService {
	return &SessionService{activityRepo: activityRepo}
}

// GetSessions 获取指定日期的工作会话列表
func (s *SessionService) GetSessions(ctx context.Context, userID int, date string) ([]*session.WorkSession, error) {
	activities, _, err := s.activityRepo.QueryByDate(ctx, userID, date, "", "", 2000, 0)
	if err != nil {
		return nil, fmt.Errorf("get sessions: %w", err)
	}

	sessions := intelligence.BuildWorkSessions(activities)

	// 转为指针切片
	result := make([]*session.WorkSession, len(sessions))
	for i := range sessions {
		result[i] = &sessions[i]
	}
	return result, nil
}

// GetIntents 获取指定日期的意图分析结果
func (s *SessionService) GetIntents(ctx context.Context, userID int, date string) (*session.IntentAnalysisResult, error) {
	activities, _, err := s.activityRepo.QueryByDate(ctx, userID, date, "", "", 2000, 0)
	if err != nil {
		return nil, fmt.Errorf("get intents: %w", err)
	}

	// 先构建会话，再分析意图
	sessions := intelligence.BuildWorkSessions(activities)
	result := intelligence.AnalyzeIntents(sessions)
	return result, nil
}

// GetTodos 获取日期范围内的待办事项
func (s *SessionService) GetTodos(ctx context.Context, userID int, from, to string) (*session.TodoExtractionResult, error) {
	// 收集范围内所有日期的活动
	allActivities, err := s.collectActivitiesInRange(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get todos: %w", err)
	}

	result := intelligence.ExtractTodos(allActivities)
	return result, nil
}

// GenerateWeekly 生成周报
func (s *SessionService) GenerateWeekly(ctx context.Context, userID int, from, to string) (*session.WeeklyReview, error) {
	// 按天收集活动
	activitiesByDay := make(map[string][]*activity.Activity)

	fromT, _ := time.Parse("2006-01-02", from)
	toT, _ := time.Parse("2006-01-02", to)

	current := fromT
	for !current.After(toT) {
		date := current.Format("2006-01-02")
		activities, _, err := s.activityRepo.QueryByDate(ctx, userID, date, "", "", 2000, 0)
		if err != nil {
			return nil, fmt.Errorf("generate weekly: %w", err)
		}
		if len(activities) > 0 {
			activitiesByDay[date] = activities
		}
		current = current.Add(24 * time.Hour)
	}

	review := intelligence.GenerateWeeklyReview(activitiesByDay, from, to)
	return review, nil
}

// collectActivitiesInRange 收集日期范围内的所有活动
func (s *SessionService) collectActivitiesInRange(ctx context.Context, userID int, from, to string) ([]*activity.Activity, error) {
	fromT, _ := time.Parse("2006-01-02", from)
	toT, _ := time.Parse("2006-01-02", to)

	var all []*activity.Activity
	current := fromT
	for !current.After(toT) {
		date := current.Format("2006-01-02")
		activities, _, err := s.activityRepo.QueryByDate(ctx, userID, date, "", "", 2000, 0)
		if err != nil {
			return nil, err
		}
		all = append(all, activities...)
		current = current.Add(24 * time.Hour)
	}
	return all, nil
}

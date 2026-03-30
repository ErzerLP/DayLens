package application

import (
	"context"
	"fmt"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/session"
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
// 实际会话聚合由 intelligence 模块完成，此处编排调用
func (s *SessionService) GetSessions(ctx context.Context, userID int, date string) ([]*session.WorkSession, error) {
	activities, _, err := s.activityRepo.QueryByDate(ctx, userID, date, "", "", 2000, 0)
	if err != nil {
		return nil, fmt.Errorf("get sessions: %w", err)
	}
	// TODO: 阶段 9 集成 intelligence.BuildSessions()
	_ = activities
	return nil, nil
}

// GetIntents 获取指定日期的意图分析
func (s *SessionService) GetIntents(ctx context.Context, userID int, date string) (*session.IntentAnalysisResult, error) {
	activities, _, err := s.activityRepo.QueryByDate(ctx, userID, date, "", "", 2000, 0)
	if err != nil {
		return nil, fmt.Errorf("get intents: %w", err)
	}
	// TODO: 阶段 9 集成 intelligence.ClassifyIntents()
	_ = activities
	return &session.IntentAnalysisResult{
		Items:          []session.IntentItem{},
		DominantIntent: "",
	}, nil
}

// GetTodos 获取日期范围内的待办事项
func (s *SessionService) GetTodos(ctx context.Context, userID int, from, to string) (*session.TodoExtractionResult, error) {
	// 将范围内每天的活动合并
	activities, _, err := s.activityRepo.QueryByDate(ctx, userID, from, "", "", 2000, 0)
	if err != nil {
		return nil, fmt.Errorf("get todos: %w", err)
	}
	// TODO: 阶段 9 集成 intelligence.ExtractTodos()
	_ = activities
	return &session.TodoExtractionResult{
		Items:   []session.TodoItem{},
		Summary: "",
	}, nil
}

// GenerateWeekly 生成周报
func (s *SessionService) GenerateWeekly(ctx context.Context, userID int, from, to string) (*session.WeeklyReview, error) {
	// TODO: 阶段 9 集成 intelligence.BuildWeeklyReview()
	return &session.WeeklyReview{
		Period:             fmt.Sprintf("%s ~ %s", from, to),
		Content:            "周报功能将在阶段 9 实现",
		DeepWorkSessions:   []session.DeepWorkSession{},
		TopApps:            []activity.AppUsage{},
		IntentDistribution: []session.IntentItem{},
	}, nil
}

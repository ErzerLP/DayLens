package application

import (
	"context"
	"fmt"
	"log/slog"

	"daylens-server/internal/application/port"
	"daylens-server/internal/domain/activity"
)

// SearchService 搜索与 AI 问答用例编排
type SearchService struct {
	activityRepo port.ActivityRepository
	aiProvider   port.AIProvider
}

// NewSearchService 创建搜索服务
func NewSearchService(activityRepo port.ActivityRepository, aiProvider port.AIProvider) *SearchService {
	return &SearchService{activityRepo: activityRepo, aiProvider: aiProvider}
}

// Search 全文搜索活动
func (s *SearchService) Search(ctx context.Context, userID int, query string, limit int) ([]*activity.SearchResultItem, int, error) {
	if query == "" {
		return nil, 0, fmt.Errorf("search: %w: query", ErrFieldMissing)
	}
	items, total, err := s.activityRepo.Search(ctx, userID, query, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}
	return items, total, nil
}

// Ask AI 问答，搜索相关活动后让 AI 回答
func (s *SearchService) Ask(ctx context.Context, userID int, question, dateContext string) (*activity.AiAnswer, error) {
	if question == "" {
		return nil, fmt.Errorf("ask: %w: question", ErrFieldMissing)
	}

	// 搜索相关活动作为上下文
	items, _, err := s.activityRepo.Search(ctx, userID, question, 10)
	if err != nil {
		slog.Warn("AI 问答搜索上下文失败", "error", err)
		items = nil
	}

	contextText := buildSearchContext(items, dateContext)
	messages := []port.Message{
		{Role: "system", Content: "你是一个工作回顾助手。根据用户的电脑使用记录回答问题。简洁准确。"},
		{Role: "user", Content: fmt.Sprintf("上下文数据:\n%s\n\n问题: %s", contextText, question)},
	}

	answer, err := s.aiProvider.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("ask ai: %w", err)
	}

	refs := make([]activity.AiReference, 0, len(items))
	for _, item := range items {
		refs = append(refs, activity.AiReference{
			ActivityID: item.ActivityID,
			Timestamp:  item.Timestamp,
			Excerpt:    item.Excerpt,
		})
	}

	return &activity.AiAnswer{Answer: answer, References: refs}, nil
}

// Chat AI 助手多轮对话
func (s *SearchService) Chat(ctx context.Context, userID int, chatMessages []activity.ChatMessage, tools []string) (*activity.AssistantReply, error) {
	messages := make([]port.Message, 0, len(chatMessages)+1)
	messages = append(messages, port.Message{
		Role:    "system",
		Content: "你是 DayLens 助手，可以分析用户的工作数据、生成建议。",
	})
	for _, m := range chatMessages {
		messages = append(messages, port.Message{Role: m.Role, Content: m.Content})
	}

	reply, err := s.aiProvider.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}

	return &activity.AssistantReply{
		Reply:     reply,
		ToolCalls: []activity.ToolCall{},
	}, nil
}

// buildSearchContext 将搜索结果构建为 AI 上下文
func buildSearchContext(items []*activity.SearchResultItem, dateContext string) string {
	if len(items) == 0 {
		return fmt.Sprintf("日期: %s\n暂无相关活动记录。", dateContext)
	}
	text := fmt.Sprintf("日期: %s\n相关活动记录:\n", dateContext)
	for _, item := range items {
		text += fmt.Sprintf("- [%s] %s: %s\n", item.AppName, item.MatchField, item.Excerpt)
	}
	return text
}

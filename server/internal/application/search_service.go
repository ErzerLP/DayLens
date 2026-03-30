package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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

// Chat AI 助手多轮对话（支持自动工具调用）
func (s *SearchService) Chat(ctx context.Context, userID int, chatMessages []activity.ChatMessage, tools []string) (*activity.AssistantReply, error) {
	// 1. 如果请求了工具，先自动执行工具调用收集上下文
	var toolCalls []activity.ToolCall
	var toolContext string

	for _, tool := range tools {
		switch tool {
		case "search":
			// 从最后一条用户消息中提取关键词作为搜索
			lastMsg := ""
			for i := len(chatMessages) - 1; i >= 0; i-- {
				if chatMessages[i].Role == "user" {
					lastMsg = chatMessages[i].Content
					break
				}
			}
			if lastMsg != "" {
				items, total, err := s.activityRepo.Search(ctx, userID, lastMsg, 10)
				if err == nil && total > 0 {
					searchResult := fmt.Sprintf("搜索到 %d 条相关活动:\n", total)
					for _, item := range items {
						searchResult += fmt.Sprintf("- [%s] %s: %s\n", item.AppName, item.MatchField, item.Excerpt)
					}
					toolCalls = append(toolCalls, activity.ToolCall{
						Tool:   "search",
						Input:  map[string]string{"query": lastMsg},
						Output: searchResult,
					})
					toolContext += "\n[搜索结果]\n" + searchResult
				}
			}
		case "stats":
			stats, err := s.activityRepo.GetDailyStats(ctx, userID, time.Now().Format("2006-01-02"))
			if err == nil && stats != nil {
				statsResult := fmt.Sprintf("今日统计: 总时长 %d 分钟, 活跃 %d 小时, 截图 %d 张\n",
					stats.TotalDuration/60, stats.ActiveHours, stats.ScreenshotCount)
				for _, app := range stats.AppUsage {
					statsResult += fmt.Sprintf("- %s: %d 分钟\n", app.AppName, app.Duration/60)
				}
				toolCalls = append(toolCalls, activity.ToolCall{
					Tool:   "stats",
					Input:  map[string]string{"date": time.Now().Format("2006-01-02")},
					Output: statsResult,
				})
				toolContext += "\n[今日统计]\n" + statsResult
			}
		}
	}

	// 2. 构建消息列表
	messages := make([]port.Message, 0, len(chatMessages)+2)
	systemPrompt := "你是 DayLens 助手，可以分析用户的工作数据、生成建议。"
	if toolContext != "" {
		systemPrompt += "\n\n以下是自动收集的参考数据：" + toolContext
	}
	messages = append(messages, port.Message{Role: "system", Content: systemPrompt})

	for _, m := range chatMessages {
		messages = append(messages, port.Message{Role: m.Role, Content: m.Content})
	}

	// 3. 调用 AI
	reply, err := s.aiProvider.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}

	return &activity.AssistantReply{
		Reply:     reply,
		ToolCalls: toolCalls,
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

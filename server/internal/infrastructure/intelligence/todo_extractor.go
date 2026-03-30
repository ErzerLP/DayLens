package intelligence

import (
	"fmt"
	"strings"
	"time"

	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/session"
)

// TODO 三级匹配模式

// todoCheckboxPatterns 复选框模式（最高置信度）
var todoCheckboxPatterns = []string{
	"[ ]", "☐", "TODO:", "TODO：", "FIXME:", "FIXME：", "HACK:",
}

// todoExplicitPatterns 显式模式（中置信度）
var todoExplicitPatterns = []string{
	"待办", "待完成", "待处理", "需要完成", "需要处理",
	"to do", "to-do", "action item", "follow up", "follow-up",
	"deadline", "截止", "到期",
}

// todoActionPatterns 动作型模式（低置信度）
var todoActionPatterns = []string{
	"需要", "应该", "必须", "记得", "别忘了",
	"should", "must", "need to", "remember to", "don't forget",
}

// ExtractTodos 从活动列表中提取待办事项
func ExtractTodos(activities []*activity.Activity) *session.TodoExtractionResult {
	var items []session.TodoItem
	seen := make(map[string]bool) // 去重

	for _, a := range activities {
		// 扫描窗口标题
		extractFromText(a.WindowTitle, a.AppName, a.Timestamp, seen, &items)

		// 扫描 OCR 文本
		if a.OcrText != nil && *a.OcrText != "" {
			lines := strings.Split(*a.OcrText, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				extractFromText(line, a.AppName+" (OCR)", a.Timestamp, seen, &items)
			}
		}
	}

	summary := fmt.Sprintf("从 %d 条活动中提取了 %d 个待办事项", len(activities), len(items))

	return &session.TodoExtractionResult{
		Items:   items,
		Summary: summary,
	}
}

// extractFromText 从单行文本中提取待办
func extractFromText(text, source string, timestamp int64, seen map[string]bool, items *[]session.TodoItem) {
	lower := strings.ToLower(text)

	// 第一优先级：复选框模式
	for _, p := range todoCheckboxPatterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			title := cleanTodoTitle(text, p)
			if title != "" && !seen[title] {
				seen[title] = true
				*items = append(*items, session.TodoItem{
					Title:       title,
					Source:      source,
					Confidence:  "high",
					ExtractedAt: timestamp,
				})
			}
			return
		}
	}

	// 第二优先级：显式模式
	for _, p := range todoExplicitPatterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			title := cleanTodoTitle(text, p)
			if title != "" && !seen[title] {
				seen[title] = true
				*items = append(*items, session.TodoItem{
					Title:       title,
					Source:      source,
					Confidence:  "medium",
					ExtractedAt: timestamp,
				})
			}
			return
		}
	}

	// 第三优先级：动作型模式
	for _, p := range todoActionPatterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			// 动作模式需要更长文本才有意义
			if len(text) < 10 {
				continue
			}
			title := text
			if !seen[title] {
				seen[title] = true
				*items = append(*items, session.TodoItem{
					Title:       title,
					Source:      source,
					Confidence:  "low",
					ExtractedAt: timestamp,
				})
			}
			return
		}
	}
}

// cleanTodoTitle 清理待办标题
func cleanTodoTitle(text string, pattern string) string {
	// 尝试从 pattern 后截取标题
	idx := strings.Index(strings.ToLower(text), strings.ToLower(pattern))
	if idx >= 0 {
		after := strings.TrimSpace(text[idx+len(pattern):])
		if after != "" {
			return after
		}
	}
	return strings.TrimSpace(text)
}

// ExtractTodosForRange 在日期范围内提取待办
func ExtractTodosForRange(activities []*activity.Activity, from, to string) *session.TodoExtractionResult {
	fromT, _ := time.Parse("2006-01-02", from)
	toT, _ := time.Parse("2006-01-02", to)
	toT = toT.Add(24 * time.Hour)

	var filtered []*activity.Activity
	for _, a := range activities {
		ts := time.Unix(a.Timestamp, 0)
		if !ts.Before(fromT) && ts.Before(toT) {
			filtered = append(filtered, a)
		}
	}

	return ExtractTodos(filtered)
}

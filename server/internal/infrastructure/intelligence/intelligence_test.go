package intelligence

import (
	"testing"

	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/session"
)

// === 会话聚合测试 ===

func TestBuildWorkSessions_EmptyInput(t *testing.T) {
	// 空列表返回 nil
	result := BuildWorkSessions(nil)
	if result != nil {
		t.Fatalf("期望 nil，实际 %v", result)
	}
}

func TestBuildWorkSessions_SingleActivity(t *testing.T) {
	// 单条活动应生成 1 个会话
	activities := []*activity.Activity{
		{Timestamp: 1000, Duration: 30, AppName: "Code", WindowTitle: "main.go"},
	}
	sessions := BuildWorkSessions(activities)
	if len(sessions) != 1 {
		t.Fatalf("期望 1 个会话，实际 %d", len(sessions))
	}
	if sessions[0].DominantApp != "Code" {
		t.Errorf("主导应用应为 Code，实际 %s", sessions[0].DominantApp)
	}
}

func TestBuildWorkSessions_SplitByGap(t *testing.T) {
	// 间隔 >15 分钟应拆分为两个会话
	activities := []*activity.Activity{
		{Timestamp: 1000, Duration: 30, AppName: "Code", WindowTitle: "a.go"},
		{Timestamp: 1030, Duration: 30, AppName: "Code", WindowTitle: "b.go"},
		{Timestamp: 3000, Duration: 30, AppName: "Chrome", WindowTitle: "Google"}, // 间隔 1940s > 900s
	}
	sessions := BuildWorkSessions(activities)
	if len(sessions) != 2 {
		t.Fatalf("期望 2 个会话，实际 %d", len(sessions))
	}
	if sessions[0].DominantApp != "Code" {
		t.Errorf("第一个会话主导应用应为 Code")
	}
	if sessions[1].DominantApp != "Chrome" {
		t.Errorf("第二个会话主导应用应为 Chrome")
	}
}

func TestBuildWorkSessions_ContinuousActivities(t *testing.T) {
	// 连续活动（间隔 <15 分钟）归入同一会话
	activities := []*activity.Activity{
		{Timestamp: 1000, Duration: 300, AppName: "Code", WindowTitle: "a.go"},
		{Timestamp: 1300, Duration: 300, AppName: "Terminal", WindowTitle: "go build"},
		{Timestamp: 1600, Duration: 300, AppName: "Code", WindowTitle: "b.go"},
	}
	sessions := BuildWorkSessions(activities)
	if len(sessions) != 1 {
		t.Fatalf("期望 1 个会话，实际 %d", len(sessions))
	}
	if len(sessions[0].Activities) != 3 {
		t.Errorf("会话应含 3 条活动，实际 %d", len(sessions[0].Activities))
	}
}

// === 意图分类测试 ===

func TestClassifySession_CodingIntent(t *testing.T) {
	// IDE 活动应识别为编程开发
	activities := []session.SessionActivity{
		{AppName: "Code", Duration: 300, Title: "main.go - Visual Studio Code"},
	}
	intent := ClassifySession(activities, nil)
	if intent.Label != IntentCoding {
		t.Errorf("期望编程开发，实际 %s", intent.Label)
	}
	if intent.Confidence <= 0 {
		t.Errorf("置信度应 > 0")
	}
}

func TestClassifySession_CommunicationIntent(t *testing.T) {
	// 微信活动应识别为沟通协作
	activities := []session.SessionActivity{
		{AppName: "微信", Duration: 300, Title: "工作群 - 微信"},
	}
	intent := ClassifySession(activities, nil)
	if intent.Label != IntentCommunication {
		t.Errorf("期望沟通协作，实际 %s", intent.Label)
	}
}

func TestClassifySession_DesignIntent(t *testing.T) {
	// Figma 应识别为设计创作
	activities := []session.SessionActivity{
		{AppName: "Figma", Duration: 300, Title: "Design System - Figma"},
	}
	intent := ClassifySession(activities, nil)
	if intent.Label != IntentDesign {
		t.Errorf("期望设计创作，实际 %s", intent.Label)
	}
}

func TestClassifySession_EmptyActivities(t *testing.T) {
	// 空活动应返回其他
	intent := ClassifySession(nil, nil)
	if intent.Label != IntentOther {
		t.Errorf("期望其他，实际 %s", intent.Label)
	}
}

// === TODO 提取测试 ===

func TestExtractTodos_CheckboxPattern(t *testing.T) {
	// 窗口标题含 TODO: 应提取
	activities := []*activity.Activity{
		{Timestamp: 1000, AppName: "Code", WindowTitle: "TODO: 修复登录bug"},
	}
	result := ExtractTodos(activities)
	if len(result.Items) != 1 {
		t.Fatalf("期望 1 个待办，实际 %d", len(result.Items))
	}
	if result.Items[0].Confidence != "high" {
		t.Errorf("TODO: 应为高置信度，实际 %s", result.Items[0].Confidence)
	}
}

func TestExtractTodos_ExplicitPattern(t *testing.T) {
	// 含"待办"关键词
	activities := []*activity.Activity{
		{Timestamp: 1000, AppName: "Notion", WindowTitle: "待办 完成API文档"},
	}
	result := ExtractTodos(activities)
	if len(result.Items) != 1 {
		t.Fatalf("期望 1 个待办，实际 %d", len(result.Items))
	}
	if result.Items[0].Confidence != "medium" {
		t.Errorf("待办应为中置信度，实际 %s", result.Items[0].Confidence)
	}
}

func TestExtractTodos_Deduplication(t *testing.T) {
	// 相同标题不应重复提取
	activities := []*activity.Activity{
		{Timestamp: 1000, AppName: "Code", WindowTitle: "TODO: 修复bug"},
		{Timestamp: 1030, AppName: "Code", WindowTitle: "TODO: 修复bug"},
	}
	result := ExtractTodos(activities)
	if len(result.Items) != 1 {
		t.Fatalf("期望去重后 1 个待办，实际 %d", len(result.Items))
	}
}

func TestExtractTodos_OcrText(t *testing.T) {
	// OCR 文本中的 TODO 也应提取
	ocrText := "第1行\nTODO: 更新文档\n第3行"
	activities := []*activity.Activity{
		{Timestamp: 1000, AppName: "Code", WindowTitle: "main.go", OcrText: &ocrText},
	}
	result := ExtractTodos(activities)
	if len(result.Items) != 1 {
		t.Fatalf("期望 OCR 提取 1 个待办，实际 %d", len(result.Items))
	}
}

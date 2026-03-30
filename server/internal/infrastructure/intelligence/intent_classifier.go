package intelligence

import (
	"strings"

	"daylens-server/internal/domain/activity"
	"daylens-server/internal/domain/session"
)

// 意图标签常量（9 种）
const (
	IntentCoding        = "编程开发"
	IntentWriting       = "文档写作"
	IntentCommunication = "沟通协作"
	IntentResearch      = "调研学习"
	IntentDesign        = "设计创作"
	IntentDataAnalysis  = "数据分析"
	IntentMeeting       = "视频会议"
	IntentBrowsing      = "网页浏览"
	IntentOther         = "其他"
)

// intentKeywords 每种意图对应的关键词及评分权重
var intentKeywords = map[string][]weightedKeyword{
	IntentCoding: {
		{"code", 10}, {"vscode", 10}, {"visual studio", 10}, {"intellij", 10}, {"idea", 8},
		{"goland", 10}, {"webstorm", 10}, {"pycharm", 10}, {"sublime", 8}, {"vim", 8},
		{"nvim", 8}, {"terminal", 6}, {"git", 7}, {"github", 6}, {"gitlab", 6},
		{".go", 5}, {".rs", 5}, {".ts", 5}, {".py", 5}, {".java", 5},
		{"debug", 6}, {"build", 5}, {"compile", 5}, {"cargo", 7}, {"npm", 6},
	},
	IntentWriting: {
		{"word", 8}, {"docs", 8}, {"notion", 9}, {"obsidian", 9}, {"typora", 9},
		{"markdown", 7}, {".md", 6}, {"文档", 8}, {"写作", 8}, {"编辑", 5},
		{"confluence", 8}, {"飞书文档", 9}, {"语雀", 9}, {"石墨", 8},
	},
	IntentCommunication: {
		{"微信", 9}, {"wechat", 9}, {"钉钉", 9}, {"dingtalk", 9}, {"slack", 9},
		{"teams", 8}, {"飞书", 9}, {"lark", 9}, {"telegram", 8}, {"discord", 8},
		{"邮件", 7}, {"mail", 7}, {"outlook", 8}, {"qq", 7},
	},
	IntentResearch: {
		{"chrome", 5}, {"firefox", 5}, {"edge", 5}, {"safari", 5},
		{"stackoverflow", 8}, {"google", 6}, {"百度", 5}, {"知乎", 7},
		{"wiki", 6}, {"documentation", 7}, {"文档", 5}, {"学习", 7},
		{"csdn", 6}, {"掘金", 6}, {"博客", 5}, {"arxiv", 8},
	},
	IntentDesign: {
		{"figma", 10}, {"sketch", 10}, {"photoshop", 10}, {"illustrator", 10},
		{"xd", 8}, {"canva", 8}, {"设计", 8}, {"design", 7},
		{"blender", 8}, {"after effects", 8},
	},
	IntentDataAnalysis: {
		{"excel", 8}, {"tableau", 10}, {"powerbi", 10}, {"jupyter", 9},
		{"数据", 6}, {"分析", 6}, {"报表", 7}, {"dashboard", 7},
		{"pandas", 8}, {"sql", 7}, {"database", 6},
	},
	IntentMeeting: {
		{"zoom", 10}, {"腾讯会议", 10}, {"飞书会议", 10}, {"会议", 9},
		{"meeting", 9}, {"webex", 9}, {"视频通话", 8},
	},
}

type weightedKeyword struct {
	keyword string
	weight  int
}

// ClassifySession 对会话进行意图分类（多关键词评分）
func ClassifySession(activities []session.SessionActivity, raw []*activity.Activity) session.IntentInfo {
	scores := make(map[string]int)
	evidenceMap := make(map[string][]string)

	for _, a := range activities {
		text := strings.ToLower(a.AppName + " " + a.Title)
		for intent, keywords := range intentKeywords {
			for _, kw := range keywords {
				if strings.Contains(text, kw.keyword) {
					scores[intent] += kw.weight
					evidence := a.AppName
					if len(evidenceMap[intent]) < 3 {
						evidenceMap[intent] = appendUnique(evidenceMap[intent], evidence)
					}
				}
			}
		}
	}

	// 用 raw 活动中的 URL 补充评分
	for _, a := range raw {
		if a.BrowserURL == nil || *a.BrowserURL == "" {
			continue
		}
		url := strings.ToLower(*a.BrowserURL)
		for intent, keywords := range intentKeywords {
			for _, kw := range keywords {
				if strings.Contains(url, kw.keyword) {
					scores[intent] += kw.weight / 2 // URL 权重减半
				}
			}
		}
	}

	if len(scores) == 0 {
		return session.IntentInfo{Label: IntentOther, Confidence: 20, Evidence: []string{}}
	}

	// 找最高分意图
	var bestIntent string
	var bestScore int
	var totalScore int
	for intent, score := range scores {
		totalScore += score
		if score > bestScore {
			bestScore = score
			bestIntent = intent
		}
	}

	// 置信度 = 最高分 / 总分 * 100
	confidence := 0
	if totalScore > 0 {
		confidence = bestScore * 100 / totalScore
	}
	if confidence > 100 {
		confidence = 100
	}

	evidence := evidenceMap[bestIntent]
	if evidence == nil {
		evidence = []string{}
	}

	return session.IntentInfo{
		Label:      bestIntent,
		Confidence: confidence,
		Evidence:   evidence,
	}
}

// AnalyzeIntents 生成意图分析结果
func AnalyzeIntents(sessions []session.WorkSession) *session.IntentAnalysisResult {
	intentDuration := make(map[string]int64)
	intentCount := make(map[string]int)
	var totalDuration int64

	for _, s := range sessions {
		intentDuration[s.Intent.Label] += s.TotalDuration
		intentCount[s.Intent.Label]++
		totalDuration += s.TotalDuration
	}

	var items []session.IntentItem
	var dominantIntent string
	var maxDuration int64

	for label, dur := range intentDuration {
		pct := 0
		if totalDuration > 0 {
			pct = int(dur * 100 / totalDuration)
		}
		items = append(items, session.IntentItem{
			Label:         label,
			TotalDuration: dur,
			SessionCount:  intentCount[label],
			Percentage:    pct,
		})
		if dur > maxDuration {
			maxDuration = dur
			dominantIntent = label
		}
	}

	return &session.IntentAnalysisResult{
		Items:                 items,
		DominantIntent:        dominantIntent,
		TotalAnalyzedDuration: totalDuration,
	}
}

// appendUnique 去重追加
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

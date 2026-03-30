// Package report 日报领域实体。
package report

// DailyReport 日报
type DailyReport struct {
	Date      string `json:"date"`
	Content   string `json:"content"`
	AIMode    string `json:"aiMode"`
	ModelName string `json:"modelName"`
	UsedAI    bool   `json:"usedAi"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

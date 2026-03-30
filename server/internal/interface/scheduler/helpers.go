package scheduler

import "daylens-server/internal/domain/activity"

// toHourlySummaryDomain 将内部结构转为领域类型
func toHourlySummaryDomain(s *struct {
	Hour                      int
	Summary                   string
	MainApps                  string
	ActivityCount             int64
	TotalDuration             int64
	RepresentativeScreenshots []string
}) *activity.HourlySummary {
	return &activity.HourlySummary{
		Hour:                      s.Hour,
		Summary:                   s.Summary,
		MainApps:                  s.MainApps,
		ActivityCount:             s.ActivityCount,
		TotalDuration:             s.TotalDuration,
		RepresentativeScreenshots: s.RepresentativeScreenshots,
	}
}

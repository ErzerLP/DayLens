package persistence

import "daylens-server/internal/application/port"

// 编译期接口满足检查
var (
	_ port.ActivityRepository     = (*PgActivityRepo)(nil)
	_ port.ReportRepository       = (*PgReportRepo)(nil)
	_ port.HourlyRepository       = (*PgHourlyRepo)(nil)
	_ port.CategoryRuleRepository = (*PgCategoryRepo)(nil)
)

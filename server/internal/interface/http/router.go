// Package http 路由注册。
package http

import (
	"github.com/gin-gonic/gin"

	"daylens-server/internal/application/port"
	"daylens-server/internal/interface/http/handler"
	"daylens-server/internal/interface/http/middleware"
)

// RouterDeps 路由注册所需依赖
type RouterDeps struct {
	Token       string
	Activity    *handler.ActivityHandler
	Stats       *handler.StatsHandler
	Report      *handler.ReportHandler
	Session     *handler.SessionHandler
	Search      *handler.SearchHandler
	App         *handler.AppHandler
	Storage     *handler.StorageHandler
	Screenshot  *handler.ScreenshotHandler
	Config      *handler.ConfigHandler
	WS          *handler.WSHandler
	Internal    *handler.InternalHandler
	HourlyRepo  port.HourlyRepository
}

// NewPublicRouter 创建公共 API 路由
func NewPublicRouter(deps *RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS(), middleware.Logger(), gin.Recovery())

	v1 := r.Group("/api/v1", middleware.Auth(deps.Token))
	{
		// 数据上报
		v1.POST("/activities", deps.Activity.Ingest)
		v1.POST("/activities/batch", deps.Activity.IngestBatch)
		v1.POST("/screenshots", deps.Screenshot.Upload)

		// 数据查询
		v1.GET("/stats", deps.Stats.GetStats)
		v1.GET("/activities", deps.Activity.List)
		v1.GET("/activities/:id", deps.Activity.Get)
		v1.GET("/hourly-summaries", deps.Stats.GetHourlySummaries)
		v1.GET("/screenshots/*key", deps.Screenshot.Get)

		// 日报
		v1.GET("/reports/:date", deps.Report.Get)
		v1.POST("/reports/generate", deps.Report.Generate)
		v1.GET("/reports/:date/export", deps.Report.Export)

		// 工作智能
		v1.GET("/sessions", deps.Session.GetSessions)
		v1.GET("/intents", deps.Session.GetIntents)
		v1.POST("/weekly-review", deps.Session.WeeklyReview)
		v1.GET("/todos", deps.Session.GetTodos)

		// 搜索 + AI
		v1.GET("/search", deps.Search.Search)
		v1.POST("/ask", deps.Search.Ask)
		v1.POST("/assistant/chat", deps.Search.Chat)

		// 应用管理
		v1.GET("/apps/recent", deps.App.Recent)
		v1.GET("/apps/categories", deps.App.Categories)
		v1.PUT("/apps/category-rules", deps.App.SetRule)
		v1.POST("/apps/reclassify", deps.App.Reclassify)

		// 存储
		v1.GET("/storage/stats", deps.Storage.Stats)
		v1.DELETE("/activities/before", deps.Storage.DeleteBefore)

		// AI 配置
		v1.GET("/config/ai", deps.Config.GetAI)
		v1.PUT("/config/ai", deps.Config.SaveAI)
		v1.POST("/config/ai/test", deps.Config.TestAI)
		v1.GET("/config/ai/providers", deps.Config.ListProviders)

		// WebSocket
		v1.GET("/ws", deps.WS.Upgrade)
	}

	return r
}

// NewInternalRouter 创建内部 API 路由（无认证）
func NewInternalRouter(internal *handler.InternalHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), gin.Recovery())

	iv1 := r.Group("/internal/v1")
	{
		iv1.GET("/health", internal.Health)
		iv1.GET("/activities", internal.Activities)
		iv1.GET("/stats", internal.Stats)
		iv1.GET("/sessions", internal.Sessions)
		iv1.GET("/intents", internal.Intents)
		iv1.GET("/reports/:date", internal.Report)
		iv1.GET("/weekly-review", internal.WeeklyReview)
		iv1.GET("/todos", internal.Todos)
		iv1.GET("/search", internal.Search)
		iv1.GET("/export", internal.Export)
	}

	return r
}

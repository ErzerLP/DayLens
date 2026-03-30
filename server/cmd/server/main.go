// Package main Composition Root — 服务启动、依赖组装
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"daylens-server/config"
	"daylens-server/internal/application"
	"daylens-server/internal/infrastructure/ai"
	"daylens-server/internal/infrastructure/crypto"
	"daylens-server/internal/infrastructure/persistence"
	"daylens-server/internal/infrastructure/storage"
	"daylens-server/internal/infrastructure/ws"
	apphttp "daylens-server/internal/interface/http"
	"daylens-server/internal/interface/http/handler"
	"daylens-server/internal/interface/scheduler"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()
	cfgJSON, _ := json.Marshal(cfg)
	slog.Info("配置加载完成", "config", string(cfgJSON))

	// ===== 基础设施层 =====
	ctx := context.Background()

	pool, err := persistence.NewPool(ctx, &cfg.Database)
	if err != nil {
		slog.Error("数据库连接失败", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 字段加密器
	var fieldCipher crypto.FieldCipher
	if cfg.Security.EncryptionKey != "" {
		keyBytes, err := hex.DecodeString(cfg.Security.EncryptionKey)
		if err != nil {
			slog.Error("加密密钥格式错误（需 64 位 hex）", "error", err)
			os.Exit(1)
		}
		fieldCipher, err = crypto.NewCipher(keyBytes)
		if err != nil {
			slog.Error("创建加密器失败", "error", err)
			os.Exit(1)
		}
		slog.Info("敏感字段加密已启用 (AES-256-GCM)")
	} else {
		fieldCipher = crypto.NopCipher{}
		slog.Warn("未配置 encryption_key，敏感字段明文存储")
	}

	activityRepo := persistence.NewPgActivityRepo(pool, fieldCipher)
	reportRepo := persistence.NewPgReportRepo(pool)
	hourlyRepo := persistence.NewPgHourlyRepo(pool)
	categoryRepo := persistence.NewPgCategoryRepo(pool)

	fileStorage := storage.NewLocalFileStorage(&cfg.Storage)
	wsHub := ws.NewHub()
	aiProvider := ai.NewProvider(&cfg.AI)

	// AI 可用性检查
	if aiProvider.IsAvailable(ctx) {
		slog.Info("AI 提供商已就绪", "provider", aiProvider.Name())
	} else {
		slog.Warn("AI 提供商不可用，日报将使用模板降级", "provider", aiProvider.Name())
	}

	// ===== 应用层 =====
	activitySvc := application.NewActivityService(activityRepo, wsHub)
	reportSvc := application.NewReportService(reportRepo, activityRepo, aiProvider, wsHub)
	sessionSvc := application.NewSessionService(activityRepo)
	searchSvc := application.NewSearchService(activityRepo, aiProvider)
	appSvc := application.NewAppService(activityRepo, categoryRepo)
	storageSvc := application.NewStorageService(activityRepo, fileStorage)

	// ===== 接口层 =====
	publicRouter := apphttp.NewPublicRouter(&apphttp.RouterDeps{
		Token:      cfg.Auth.Token,
		Activity:   handler.NewActivityHandler(activitySvc),
		Stats:      handler.NewStatsHandler(activitySvc, hourlyRepo),
		Report:     handler.NewReportHandler(reportSvc),
		Session:    handler.NewSessionHandler(sessionSvc),
		Search:     handler.NewSearchHandler(searchSvc),
		App:        handler.NewAppHandler(appSvc),
		Storage:    handler.NewStorageHandler(storageSvc),
		Screenshot: handler.NewScreenshotHandler(fileStorage),
		Config:     handler.NewConfigHandler(cfg, nil),
		WS:         handler.NewWSHandler(wsHub),
		Internal:   handler.NewInternalHandler(activitySvc, reportSvc, sessionSvc, searchSvc),
	})

	internalRouter := apphttp.NewInternalRouter(
		handler.NewInternalHandler(activitySvc, reportSvc, sessionSvc, searchSvc),
	)

	// ===== 定时任务 =====
	sched := scheduler.New(cfg, reportSvc, storageSvc, hourlyRepo, activityRepo)
	sched.Start()

	// ===== 启动 =====
	publicSrv := &http.Server{Addr: cfg.Server.Addr, Handler: publicRouter}
	internalSrv := &http.Server{Addr: cfg.Server.InternalAddr, Handler: internalRouter}

	go func() {
		slog.Info("公共 API 已启动", "addr", cfg.Server.Addr)
		if err := publicSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("公共 API 异常退出", "error", err)
		}
	}()
	go func() {
		slog.Info("内部 API 已启动", "addr", cfg.Server.InternalAddr)
		if err := internalSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("内部 API 异常退出", "error", err)
		}
	}()

	fmt.Printf("✅ DayLens Server 已就绪\n")
	fmt.Printf("   公共 API:  http://%s\n", cfg.Server.Addr)
	fmt.Printf("   内部 API:  http://%s\n", cfg.Server.InternalAddr)
	fmt.Printf("   WebSocket: ws://%s/api/v1/ws\n", cfg.Server.Addr)
	fmt.Printf("   AI:        %s\n", aiProvider.Name())

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("正在关闭服务...")
	sched.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = publicSrv.Shutdown(shutdownCtx)
	_ = internalSrv.Shutdown(shutdownCtx)
	slog.Info("服务已关闭")
}

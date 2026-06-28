// Command worker is the background job processor (Asynq). It runs as a separate
// process so workers scale independently of the HTTP API (plan §2, §3.3d).
// Queues are prioritized: critical (payments) > default (notifications) > low
// (reindex, analytics).
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/internal/jobs"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/search"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/pkg/logger"
)

var version = "dev"

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", slog.Any("err", err))
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)
	log.Info("starting worker", slog.String("env", string(cfg.App.Env)), slog.String("version", version))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Dependencies the task handlers need.
	db, err := repository.NewDB(cfg.DB, cfg.App.IsProduction())
	if err != nil {
		log.Error("db init failed", slog.Any("err", err))
		os.Exit(1)
	}
	defer repository.Close(db, log)

	catalogRepo := repository.NewCatalogRepo(db)
	translationRepo := repository.NewTranslationRepo(db)
	couponRepo := repository.NewCouponRepo(db)
	orderRepo := repository.NewOrderRepo(db, couponRepo)
	pushRepo := repository.NewPushRepo(db)
	userRepo := repository.NewUserRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)
	meili := search.NewClient(cfg.Meili)
	searchSvc := services.NewSearchService(meili, catalogRepo, translationRepo, cfg.I18n.DefaultLocale, cfg.Media.PublicBaseURL)
	pushSvc := services.NewPushService(pushRepo, cfg.VAPID)
	metaSvc := services.NewMetaService(settingsRepo, userRepo, orderRepo)

	// Asynq server with prioritized queues.
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB},
		asynq.Config{
			Concurrency: 30, // managed goroutine pool (plan §3.3d)
			Queues:      map[string]int{jobs.QueueCritical: 6, jobs.QueueDefault: 3, jobs.QueueLow: 1},
		},
	)

	mux := asynq.NewServeMux()
	jobs.NewHandlers(searchSvc, pushSvc, metaSvc, orderRepo, log).Register(mux)

	// Run until a shutdown signal arrives, then stop gracefully.
	go func() {
		<-ctx.Done()
		log.Info("worker shutting down")
		srv.Shutdown()
	}()

	if err := srv.Run(mux); err != nil {
		log.Error("worker exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

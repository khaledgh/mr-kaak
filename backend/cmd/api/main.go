// Command api is the HTTP server entrypoint. It loads config, wires up
// dependencies (DB, Redis, logger), starts the server, and shuts down
// gracefully on SIGINT/SIGTERM.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mrkaak/restaurant-api/internal/cache"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/server"
	"github.com/mrkaak/restaurant-api/pkg/logger"
)

// version is overridable at build time: -ldflags "-X main.version=$(git rev-parse --short HEAD)".
var version = "dev"

func main() {
	// Load .env if present (no-op in prod where env is injected by the platform).
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		// Logger not built yet; use the default before exiting.
		slog.Error("config load failed", slog.Any("err", err))
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)
	log.Info("starting api",
		slog.String("env", string(cfg.App.Env)),
		slog.String("version", version),
	)

	// Cancel the root context on SIGINT/SIGTERM to trigger graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := repository.NewDB(cfg.DB, cfg.App.IsProduction())
	if err != nil {
		log.Error("db init failed", slog.Any("err", err))
		os.Exit(1)
	}
	defer repository.Close(db, log)

	if err := repository.WaitForDB(ctx, db, log); err != nil {
		log.Error("db never became ready", slog.Any("err", err))
		os.Exit(1)
	}

	rdb := cache.New(cfg.Redis)
	defer func() { _ = rdb.Close() }()

	srv := server.New(server.Deps{
		Config:  cfg,
		Logger:  log,
		DB:      db,
		Redis:   rdb,
		Version: version,
	})

	if err := srv.Run(ctx); err != nil {
		log.Error("server exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

// Package server assembles the Echo HTTP server: error handling, middleware,
// route registration, and graceful shutdown.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/auth"
	"github.com/mrkaak/restaurant-api/internal/cache"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/internal/handlers"
	"github.com/mrkaak/restaurant-api/internal/jobs"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/payment"
	"github.com/mrkaak/restaurant-api/internal/realtime"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/search"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/storage"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/logger"
	"github.com/mrkaak/restaurant-api/pkg/response"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Deps are the runtime dependencies the server needs to build its handlers.
type Deps struct {
	Config  *config.Config
	Logger  *slog.Logger
	DB      *gorm.DB
	Redis   *redis.Client
	Version string
}

// Server wraps the Echo instance and its HTTP server.
type Server struct {
	echo *echo.Echo
	http *http.Server
	cfg  *config.Config
	log  *slog.Logger
}

// New builds a fully-configured server (middleware + routes) ready to Run.
func New(d Deps) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = errorHandler(d.Logger)

	middleware.Register(e, d.Config, d.Logger, d.Redis)

	// Serve uploaded files at /uploads/* (must be before registerRoutes so the
	// path pattern does not conflict with the API prefix).
	e.Static("/uploads", d.Config.Media.UploadDir)

	registerRoutes(e, d)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", d.Config.HTTP.Port),
		Handler:      e,
		ReadTimeout:  d.Config.HTTP.ReadTimeout,
		WriteTimeout: d.Config.HTTP.WriteTimeout,
		IdleTimeout:  d.Config.HTTP.IdleTimeout,
	}

	return &Server{echo: e, http: srv, cfg: d.Config, log: d.Logger}
}

// registerRoutes builds the dependency graph (repos → services → handlers) and
// mounts every route group. New feature areas extend this as the build
// progresses through the phases.
func registerRoutes(e *echo.Echo, d Deps) {
	root := e.Group("")
	handlers.NewHealth(d.DB, d.Redis, d.Version).Register(root)

	api := e.Group("/api/v1")

	// Shared primitives.
	jwtManager := auth.NewManager(d.Config.JWT)
	v := validator.New()

	// Middleware factories.
	jwtAuth := middleware.Auth(jwtManager)
	adminOnly := middleware.RequireRole(models.RoleAdmin)
	staffOnly := middleware.RequireStaff()

	// Cache (cache-aside; degrades gracefully when Redis is down).
	c := cache.NewCache(d.Redis)

	// Search (Meilisearch) + job enqueuer (Asynq). Both degrade gracefully:
	// search falls back to the DB; enqueue failures are logged, not fatal.
	meili := search.NewClient(d.Config.Meili)
	enqueuer := jobs.NewEnqueuer(d.Config.Redis)

	// Storage (local disk; swappable for S3 later).
	store, err := storage.NewLocal(d.Config.Media.UploadDir, d.Config.Media.PublicBaseURL)
	if err != nil {
		d.Logger.Error("failed to initialize local storage", slog.Any("err", err))
		panic(err)
	}

	// Repositories.
	userRepo := repository.NewUserRepo(d.DB)
	addressRepo := repository.NewAddressRepo(d.DB)
	catalogRepo := repository.NewCatalogRepo(d.DB)
	translationRepo := repository.NewTranslationRepo(d.DB)
	languageRepo := repository.NewLanguageRepo(d.DB)
	zoneRepo := repository.NewZoneRepo(d.DB)
	couponRepo := repository.NewCouponRepo(d.DB)
	bannerRepo := repository.NewBannerRepo(d.DB)
	settingsRepo := repository.NewSettingsRepo(d.DB)
	orderRepo := repository.NewOrderRepo(d.DB, couponRepo)
	paymentRepo := repository.NewPaymentRepo(d.DB)
	pushRepo := repository.NewPushRepo(d.DB)

	// Realtime hub + Redis fan-out publisher. The bridge republishes events
	// from other replicas into this instance's hub (no-op if Redis is down).
	hub := realtime.NewHub()
	publisher := realtime.NewPublisher(hub, d.Redis)
	go realtime.RunBridge(context.Background(), hub, d.Redis, d.Logger)

	// Services.
	authSvc := services.NewAuthService(userRepo, jwtManager)
	userSvc := services.NewUserService(userRepo, addressRepo)
	catalogSvc := services.NewCatalogService(catalogRepo, translationRepo, c, enqueuer, d.Config.I18n.DefaultLocale)
	searchSvc := services.NewSearchService(meili, catalogRepo, translationRepo, d.Config.I18n.DefaultLocale)
	languageSvc := services.NewLanguageService(languageRepo)
	zoneSvc := services.NewZoneService(zoneRepo)
	couponSvc := services.NewCouponService(couponRepo)
	bannerSvc := services.NewBannerService(bannerRepo)
	orderSvc := services.NewOrderService(orderRepo, catalogRepo, translationRepo, addressRepo,
		settingsRepo, zoneSvc, couponSvc, d.Config.I18n.DefaultLocale, d.Config.I18n.Currency).
		WithRealtime(publisher, enqueuer)
	paymentSvc := services.NewPaymentService(orderRepo, paymentRepo, settingsRepo, d.Config.I18n.Currency,
		payment.NewCOD(), payment.NewSquare())
	settingsSvc := services.NewSettingsService(settingsRepo)
	pushSvc := services.NewPushService(pushRepo, d.Config.VAPID)
	metaSvc := services.NewMetaService(settingsRepo, userRepo, orderRepo)

	// Handlers.
	handlers.NewAuthHandler(authSvc, userSvc, v).Register(api, jwtAuth)
	handlers.NewUserHandler(userSvc, v).Register(api, jwtAuth, adminOnly)
	handlers.NewCatalogHandler(catalogSvc, v).Register(api, jwtAuth, adminOnly)
	handlers.NewLanguageHandler(languageSvc, v).Register(api, jwtAuth, adminOnly)
	handlers.NewZoneHandler(zoneSvc, v).Register(api, jwtAuth, adminOnly)
	handlers.NewCouponHandler(couponSvc, v).Register(api, jwtAuth, adminOnly)
	handlers.NewBannerHandler(bannerSvc, v).Register(api, jwtAuth, adminOnly)
	handlers.NewOrderHandler(orderSvc, v).Register(api, jwtAuth, staffOnly)
	handlers.NewPaymentHandler(paymentSvc, v).Register(api, jwtAuth, staffOnly)
	handlers.NewSettingsHandler(settingsSvc).Register(api, jwtAuth, adminOnly)
	handlers.NewSearchHandler(searchSvc).Register(api, jwtAuth, adminOnly)
	handlers.NewStreamHandler(orderSvc, hub).Register(api, jwtAuth)
	handlers.NewPushHandler(pushSvc, v).Register(api, jwtAuth)
	handlers.NewMetaHandler(metaSvc).Register(api)

	// Media subsystem.
	mediaRepo := repository.NewMediaRepo(d.DB)
	mediaSvc := services.NewMediaService(mediaRepo, store,
		d.Config.Media.MaxBytes, d.Config.Media.AllowedMIME)
	handlers.NewMediaHandler(mediaSvc).Register(api, jwtAuth, adminOnly)
}

// Run starts the server and blocks until ctx is cancelled (e.g. SIGTERM),
// then drains in-flight requests within the configured shutdown timeout.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("http server listening", slog.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
		s.log.Info("shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.App.ShutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	s.log.Info("http server stopped cleanly")
	return nil
}

// errorHandler converts any error returned by a handler into the standard
// response envelope. Unexpected errors are logged and reported as a generic
// 500 so internal detail never leaks to clients.
func errorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var he *echo.HTTPError
		if errors.As(err, &he) {
			code := response.CodeBadRequest
			switch he.Code {
			case http.StatusUnauthorized:
				code = response.CodeUnauthorized
			case http.StatusForbidden:
				code = response.CodeForbidden
			case http.StatusNotFound:
				code = response.CodeNotFound
			case http.StatusTooManyRequests:
				code = response.CodeRateLimited
			case http.StatusConflict:
				code = response.CodeConflict
			case http.StatusUnprocessableEntity:
				code = response.CodeUnprocessable
			case http.StatusInternalServerError:
				code = response.CodeInternal
			}
			msg := http.StatusText(he.Code)
			if m, ok := he.Message.(string); ok && m != "" {
				msg = m
			}
			_ = response.Error(c, he.Code, code, msg)
			return
		}

		logger.FromContext(c.Request().Context()).Error("unhandled error",
			slog.Any("err", err), slog.String("path", c.Path()))
		_ = response.Internal(c)
	}
}

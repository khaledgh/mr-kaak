// Package middleware wires cross-cutting HTTP concerns: request id, structured
// access logging, panic recovery, CORS, body limits, timeouts, and rate
// limiting. Registration order matters and is documented in Register.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/pkg/logger"
	"github.com/mrkaak/restaurant-api/pkg/response"
	"github.com/redis/go-redis/v9"
)

// Register installs the global middleware stack on e, outermost first.
func Register(e *echo.Echo, cfg *config.Config, log *slog.Logger, rdb *redis.Client) {
	// 1. Recover from panics first so nothing below can crash the process.
	e.Use(Recover(log))

	// 2. Assign/propagate a request id for correlation across logs.
	e.Use(emw.RequestIDWithConfig(emw.RequestIDConfig{TargetHeader: echo.HeaderXRequestID}))

	// 3. Attach a request-scoped logger and emit one structured access log line.
	e.Use(RequestLogger(log))

	// 4. CORS — driven by configured origins.
	e.Use(emw.CORSWithConfig(emw.CORSConfig{
		AllowOrigins:     cfg.HTTP.AllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization, "Accept-Language", "Idempotency-Key", echo.HeaderXRequestID},
		AllowCredentials: true,
		MaxAge:           600,
	}))

	// 5. Reject oversized bodies before they hit handlers.
	e.Use(emw.BodyLimit("1M"))

	// 6. Per-instance rate limit. A Redis-backed sliding window replaces this
	//    in the hardening phase so limits are shared across replicas (§3.3f).
	e.Use(emw.RateLimiterWithConfig(emw.RateLimiterConfig{
		Store: emw.NewRateLimiterMemoryStore(100), // 100 req/s steady, burst handled by store
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited, "rate limit exceeded")
		},
		DenyHandler: func(c echo.Context, id string, err error) error {
			return response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited, "rate limit exceeded")
		},
	}))

	// 7. Stricter, Redis-backed limit on abuse-prone endpoints (auth, orders,
	//    coupon validation), shared across replicas. Fails open if Redis is down.
	if rdb != nil {
		e.Use(SensitiveRateLimit(rdb, 20, time.Minute))
	}
}

// RequestLogger attaches a request-scoped logger (with request id) to the
// context and logs one structured line per request with status and latency.
func RequestLogger(base *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			reqID := req.Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = c.Response().Header().Get(echo.HeaderXRequestID)
			}

			reqLog := base.With(slog.String("request_id", reqID))
			c.SetRequest(req.WithContext(logger.WithContext(req.Context(), reqLog)))

			err := next(c)
			if err != nil {
				// Let Echo's HTTPErrorHandler translate it, but record it here.
				c.Error(err)
			}

			level := slog.LevelInfo
			status := c.Response().Status
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}

			reqLog.LogAttrs(req.Context(), level, "http_request",
				slog.String("method", req.Method),
				slog.String("path", c.Path()),
				slog.Int("status", status),
				slog.String("ip", c.RealIP()),
				slog.Duration("latency", time.Since(start)),
			)
			return nil // error already handled above
		}
	}
}

// Recover converts panics into a 500 envelope and logs the stack.
func Recover(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.FromContext(c.Request().Context()).Error("panic recovered",
						slog.Any("panic", r),
						slog.String("path", c.Path()),
					)
					_ = response.Internal(c)
				}
			}()
			return next(c)
		}
	}
}

// Timeout enforces a hard deadline on the request context so slow downstream
// calls (DB, Redis) get cancelled instead of piling up (§3.3b).
func Timeout(d time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), d)
			defer cancel()
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

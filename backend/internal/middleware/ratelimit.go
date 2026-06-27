package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/pkg/response"
	"github.com/redis/go-redis/v9"
)

// SensitiveRateLimit applies a stricter, Redis-backed fixed-window rate limit to
// abuse-prone endpoints (auth + order placement) — keyed by client IP so the
// limit is shared across API replicas (plan §3.3f). It complements the looser
// global in-process limiter.
//
// Fail-open: if Redis is unavailable the request is allowed (the global
// in-process limiter remains as a backstop), so a Redis blip never locks users
// out. Read-heavy GETs are not limited here.
func SensitiveRateLimit(rdb *redis.Client, limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isSensitive(c) {
				return next(c)
			}
			key := fmt.Sprintf("rl:%s:%s:%d", bucket(c), c.RealIP(), time.Now().Unix()/int64(window.Seconds()))

			ctx, cancel := context.WithTimeout(c.Request().Context(), 200*time.Millisecond)
			defer cancel()

			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				return next(c) // fail-open on Redis error
			}
			if count == 1 {
				_ = rdb.Expire(ctx, key, window).Err()
			}
			if count > int64(limit) {
				return response.Error(c, http.StatusTooManyRequests, response.CodeRateLimited,
					"too many requests; please slow down")
			}
			return next(c)
		}
	}
}

// isSensitive reports whether the route should be tightly limited.
func isSensitive(c echo.Context) bool {
	p := c.Path()
	switch {
	case strings.Contains(p, "/auth/login"),
		strings.Contains(p, "/auth/register"),
		strings.Contains(p, "/auth/refresh"),
		strings.Contains(p, "/coupons/validate"):
		return true
	case strings.HasSuffix(p, "/orders") && c.Request().Method == http.MethodPost:
		return true
	default:
		return false
	}
}

// bucket groups related routes under one limit namespace.
func bucket(c echo.Context) string {
	p := c.Path()
	if strings.Contains(p, "/auth/") {
		return "auth"
	}
	if strings.HasSuffix(p, "/orders") {
		return "orders"
	}
	return "misc"
}

// Package handlers holds the Echo HTTP handlers. Handlers stay thin: parse
// input, call a service, write a response. This file exposes liveness and
// readiness checks used by load balancers and orchestrators.
package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/cache"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/pkg/response"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Health checks the API and its critical dependencies.
type Health struct {
	db      *gorm.DB
	rdb     *redis.Client
	version string
}

func NewHealth(db *gorm.DB, rdb *redis.Client, version string) *Health {
	return &Health{db: db, rdb: rdb, version: version}
}

// Register mounts the health routes on g (typically the root group).
func (h *Health) Register(g *echo.Group) {
	g.GET("/healthz", h.Live) // liveness: is the process up?
	g.GET("/readyz", h.Ready) // readiness: can it serve traffic (deps ok)?
}

// Live always returns 200 if the process is running.
func (h *Health) Live(c echo.Context) error {
	return response.OK(c, echo.Map{"status": "ok", "version": h.version})
}

// Ready pings every critical dependency and reports per-component status.
// Returns 503 if any dependency is down so the LB stops routing here.
func (h *Health) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{}
	healthy := true

	if err := repository.Ping(ctx, h.db); err != nil {
		checks["mysql"] = "down: " + err.Error()
		healthy = false
	} else {
		checks["mysql"] = "ok"
	}

	if err := cache.Ping(ctx, h.rdb); err != nil {
		checks["redis"] = "down: " + err.Error()
		healthy = false
	} else {
		checks["redis"] = "ok"
	}

	status := http.StatusOK
	state := "ok"
	if !healthy {
		status = http.StatusServiceUnavailable
		state = "degraded"
	}
	return c.JSON(status, response.Envelope{Data: echo.Map{
		"status":     state,
		"version":    h.version,
		"components": checks,
	}})
}

package handlers

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// SettingsHandler exposes the admin settings editor (payment toggles, keys,
// tax %, Meta keys, ...).
type SettingsHandler struct {
	settings *services.SettingsService
}

func NewSettingsHandler(s *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{settings: s}
}

func (h *SettingsHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	g := api.Group("/admin/settings", jwtAuth, adminOnly)
	g.GET("", h.All)
	g.PUT("", h.Update)
}

func (h *SettingsHandler) All(c echo.Context) error {
	all, err := h.settings.All(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, all)
}

func (h *SettingsHandler) Update(c echo.Context) error {
	var body map[string]json.RawMessage
	if err := c.Bind(&body); err != nil || len(body) == 0 {
		return response.BadRequest(c, "expected a JSON object of settings")
	}
	if err := h.settings.Update(c.Request().Context(), body); err != nil {
		return mapServiceError(c, err)
	}
	return h.All(c)
}

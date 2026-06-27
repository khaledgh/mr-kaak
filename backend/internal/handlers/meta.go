package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// MetaHandler exposes the public Pixel config so the frontend can initialize
// the Meta Pixel only when an ID is configured (plan §14).
type MetaHandler struct {
	meta *services.MetaService
}

func NewMetaHandler(m *services.MetaService) *MetaHandler {
	return &MetaHandler{meta: m}
}

func (h *MetaHandler) Register(api *echo.Group) {
	api.GET("/meta/config", h.Config)
}

func (h *MetaHandler) Config(c echo.Context) error {
	return response.OK(c, h.meta.Config(c.Request().Context()))
}

package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// PushHandler exposes the VAPID public key and subscribe/unsubscribe endpoints.
type PushHandler struct {
	push *services.PushService
	v    *validator.Validator
}

func NewPushHandler(p *services.PushService, v *validator.Validator) *PushHandler {
	return &PushHandler{push: p, v: v}
}

func (h *PushHandler) Register(api *echo.Group, jwtAuth echo.MiddlewareFunc) {
	api.GET("/push/vapid-public-key", h.PublicKey)
	g := api.Group("/push", jwtAuth)
	g.POST("/subscribe", h.Subscribe)
	g.POST("/unsubscribe", h.Unsubscribe)
}

func (h *PushHandler) PublicKey(c echo.Context) error {
	return response.OK(c, echo.Map{"public_key": h.push.PublicKey(), "enabled": h.push.Enabled()})
}

func (h *PushHandler) Subscribe(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	var in services.PushSubscribeInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	if err := h.push.Subscribe(c.Request().Context(), uid, in); err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, echo.Map{"subscribed": true})
}

func (h *PushHandler) Unsubscribe(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	var in services.PushUnsubscribeInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	if err := h.push.Unsubscribe(c.Request().Context(), uid, in.Endpoint); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

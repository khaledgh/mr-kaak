package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/pagination"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// OrderHandler exposes customer checkout/tracking and the staff order board.
type OrderHandler struct {
	orders *services.OrderService
	v      *validator.Validator
}

func NewOrderHandler(o *services.OrderService, v *validator.Validator) *OrderHandler {
	return &OrderHandler{orders: o, v: v}
}

func (h *OrderHandler) Register(api *echo.Group, jwtAuth, staffOnly echo.MiddlewareFunc) {
	// Customer
	g := api.Group("/orders", jwtAuth)
	g.POST("", h.Checkout)
	g.GET("", h.ListMine)
	g.GET("/:code", h.GetByCode)

	// Staff/kitchen board
	admin := api.Group("/admin/orders", jwtAuth, staffOnly)
	admin.GET("", h.AdminList)
	admin.GET("/:code", h.AdminGet)
	admin.POST("/:id/status", h.AdvanceStatus)
}

func (h *OrderHandler) Checkout(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	var in services.CheckoutInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	// Idempotency-Key header makes a retried checkout safe (plan §3.3g).
	idem := c.Request().Header.Get("Idempotency-Key")
	o, err := h.orders.Checkout(c.Request().Context(), uid, idem, in)
	if err != nil {
		return mapOrderError(c, err)
	}
	return response.Created(c, o)
}

func (h *OrderHandler) ListMine(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	p := pagination.FromQuery(c)
	orders, total, err := h.orders.ListMine(c.Request().Context(), uid, p.Limit(), p.Offset())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Paginated(c, http.StatusOK, orders, pagination.NewMeta(p, total))
}

func (h *OrderHandler) GetByCode(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	o, err := h.orders.GetByCode(c.Request().Context(), uid, false, c.Param("code"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, o)
}

func (h *OrderHandler) AdminList(c echo.Context) error {
	p := pagination.FromQuery(c)
	orders, total, err := h.orders.AdminList(c.Request().Context(), c.QueryParam("status"), p.Limit(), p.Offset())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Paginated(c, http.StatusOK, orders, pagination.NewMeta(p, total))
}

func (h *OrderHandler) AdminGet(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	o, err := h.orders.GetByCode(c.Request().Context(), uid, true, c.Param("code"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, o)
}

func (h *OrderHandler) AdvanceStatus(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.AdvanceStatusInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	o, err := h.orders.AdvanceStatus(c.Request().Context(), id, models.OrderStatus(in.Status), uid, in.Note)
	if err != nil {
		return mapOrderError(c, err)
	}
	return response.OK(c, o)
}

// mapOrderError translates order/checkout domain errors to HTTP responses.
func mapOrderError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, services.ErrItemUnavailable),
		errors.Is(err, services.ErrVariantInvalid),
		errors.Is(err, services.ErrWeightRequired),
		errors.Is(err, services.ErrModifierInvalid),
		errors.Is(err, services.ErrModifierBounds),
		errors.Is(err, services.ErrAddressNoGeo),
		errors.Is(err, services.ErrInvalidTransition):
		return response.Error(c, http.StatusUnprocessableEntity, response.CodeValidation, err.Error())
	case errors.Is(err, services.ErrAddressRequired),
		errors.Is(err, services.ErrUndeliverable),
		errors.Is(err, services.ErrBelowMinOrder),
		errors.Is(err, services.ErrPaymentMethodDisabled):
		return response.Error(c, http.StatusUnprocessableEntity, response.CodeUnprocessable, err.Error())
	case errors.Is(err, services.ErrCouponInvalid),
		errors.Is(err, services.ErrCouponExpired),
		errors.Is(err, services.ErrCouponNotStarted),
		errors.Is(err, services.ErrCouponMinOrder),
		errors.Is(err, services.ErrCouponExhausted),
		errors.Is(err, services.ErrCouponPerUser):
		return response.Error(c, http.StatusUnprocessableEntity, response.CodeValidation, err.Error())
	default:
		return mapServiceError(c, err)
	}
}

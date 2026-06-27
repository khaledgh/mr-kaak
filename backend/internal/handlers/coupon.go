package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// CouponHandler exposes the checkout coupon preview and admin coupon CRUD.
type CouponHandler struct {
	coupons *services.CouponService
	v       *validator.Validator
}

func NewCouponHandler(c *services.CouponService, v *validator.Validator) *CouponHandler {
	return &CouponHandler{coupons: c, v: v}
}

func (h *CouponHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	// Authenticated: preview a coupon's effect for the current cart.
	api.POST("/coupons/validate", h.Validate, jwtAuth)

	admin := api.Group("/admin/coupons", jwtAuth, adminOnly)
	admin.GET("", h.List)
	admin.POST("", h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)
}

func (h *CouponHandler) Validate(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	var in services.ValidateCouponInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	d, err := h.coupons.Validate(c.Request().Context(), in.Code, uid, in.SubtotalCents, in.DeliveryFeeCents)
	if err != nil {
		return mapCouponError(c, err)
	}
	return response.OK(c, d)
}

func (h *CouponHandler) List(c echo.Context) error {
	cs, err := h.coupons.List(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, cs)
}

func (h *CouponHandler) Create(c echo.Context) error {
	var in services.CouponInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	cp, err := h.coupons.Create(c.Request().Context(), in)
	if err != nil {
		return mapCatalogError(c, err) // reuse slug/code conflict mapping
	}
	return response.Created(c, cp)
}

func (h *CouponHandler) Update(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.CouponInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	cp, err := h.coupons.Update(c.Request().Context(), id, in)
	if err != nil {
		return mapCatalogError(c, err)
	}
	return response.OK(c, cp)
}

func (h *CouponHandler) Delete(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.coupons.Delete(c.Request().Context(), id); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// mapCouponError surfaces the specific coupon rejection reason as a 422 so the
// checkout UI can show why a code didn't apply.
func mapCouponError(c echo.Context, err error) error {
	switch {
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

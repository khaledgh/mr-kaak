package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/geo"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// ZoneHandler exposes the public delivery quote and admin zone CRUD.
type ZoneHandler struct {
	zones *services.ZoneService
	v     *validator.Validator
}

func NewZoneHandler(z *services.ZoneService, v *validator.Validator) *ZoneHandler {
	return &ZoneHandler{zones: z, v: v}
}

func (h *ZoneHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	// Public: check deliverability + fee for an address before checkout.
	api.POST("/delivery/quote", h.Quote)

	admin := api.Group("/admin/delivery-zones", jwtAuth, adminOnly)
	admin.GET("", h.List)
	admin.POST("", h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)
}

func (h *ZoneHandler) Quote(c echo.Context) error {
	var in services.QuoteInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	q, err := h.zones.Quote(c.Request().Context(), geo.Point{Lat: in.Lat, Lng: in.Lng}, in.ProductIDs)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, q)
}

func (h *ZoneHandler) List(c echo.Context) error {
	zones, err := h.zones.List(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, zones)
}

func (h *ZoneHandler) Create(c echo.Context) error {
	var in services.ZoneInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	z, err := h.zones.Create(c.Request().Context(), in)
	if err != nil {
		return mapZoneError(c, err)
	}
	return response.Created(c, z)
}

func (h *ZoneHandler) Update(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.ZoneInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	z, err := h.zones.Update(c.Request().Context(), id, in)
	if err != nil {
		return mapZoneError(c, err)
	}
	return response.OK(c, z)
}

func (h *ZoneHandler) Delete(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.zones.Delete(c.Request().Context(), id); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

func mapZoneError(c echo.Context, err error) error {
	if errors.Is(err, services.ErrInvalidZone) {
		return response.Error(c, http.StatusUnprocessableEntity, response.CodeValidation,
			"invalid zone: radius zones need center+radius_km, polygon zones need >=3 points")
	}
	return mapServiceError(c, err)
}

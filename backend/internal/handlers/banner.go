package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// BannerHandler exposes the public active-banners feed and admin banner CRUD.
type BannerHandler struct {
	banners *services.BannerService
	v       *validator.Validator
}

func NewBannerHandler(b *services.BannerService, v *validator.Validator) *BannerHandler {
	return &BannerHandler{banners: b, v: v}
}

func (h *BannerHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	api.GET("/banners", h.ActiveNow) // public hero carousel

	admin := api.Group("/admin/banners", jwtAuth, adminOnly)
	admin.GET("", h.List)
	admin.POST("", h.Create)
	admin.PUT("/:id", h.Update)
	admin.DELETE("/:id", h.Delete)
}

func (h *BannerHandler) ActiveNow(c echo.Context) error {
	bs, err := h.banners.ActiveNow(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, bs)
}

func (h *BannerHandler) List(c echo.Context) error {
	bs, err := h.banners.List(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, bs)
}

func (h *BannerHandler) Create(c echo.Context) error {
	var in services.BannerInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	b, err := h.banners.Create(c.Request().Context(), in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, b)
}

func (h *BannerHandler) Update(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.BannerInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	b, err := h.banners.Update(c.Request().Context(), id, in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, b)
}

func (h *BannerHandler) Delete(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.banners.Delete(c.Request().Context(), id); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

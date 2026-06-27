package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// LanguageHandler exposes public locale discovery + bundles and admin language
// management.
type LanguageHandler struct {
	langs *services.LanguageService
	v     *validator.Validator
}

func NewLanguageHandler(l *services.LanguageService, v *validator.Validator) *LanguageHandler {
	return &LanguageHandler{langs: l, v: v}
}

func (h *LanguageHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	// Public: the frontend reads active languages and loads its string bundle.
	api.GET("/languages", h.ListActive)
	api.GET("/i18n/:locale", h.Bundle)

	admin := api.Group("/admin/languages", jwtAuth, adminOnly)
	admin.GET("", h.ListAll)
	admin.POST("", h.Create)
	admin.PUT("/:id", h.Update)
	admin.POST("/:id/default", h.SetDefault)
	admin.PUT("/:locale/bundle", h.UpsertBundle)
}

func (h *LanguageHandler) ListActive(c echo.Context) error {
	langs, err := h.langs.ListActive(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, langs)
}

func (h *LanguageHandler) Bundle(c echo.Context) error {
	b, err := h.langs.Bundle(c.Request().Context(), c.Param("locale"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, b)
}

func (h *LanguageHandler) ListAll(c echo.Context) error {
	langs, err := h.langs.ListAll(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, langs)
}

func (h *LanguageHandler) Create(c echo.Context) error {
	var in services.LanguageInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	l, err := h.langs.Create(c.Request().Context(), in)
	if err != nil {
		return mapCatalogError(c, err) // reuses slug-conflict mapping for code clashes
	}
	return response.Created(c, l)
}

func (h *LanguageHandler) Update(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.LanguageInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	l, err := h.langs.Update(c.Request().Context(), id, in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, l)
}

func (h *LanguageHandler) SetDefault(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.langs.SetDefault(c.Request().Context(), id); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

func (h *LanguageHandler) UpsertBundle(c echo.Context) error {
	var in services.BundleInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	if err := h.langs.UpsertBundle(c.Request().Context(), c.Param("locale"), in.Strings); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

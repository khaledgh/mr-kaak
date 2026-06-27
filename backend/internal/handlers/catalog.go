package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// CatalogHandler serves the public menu/category/product endpoints and the
// admin catalog CRUD.
type CatalogHandler struct {
	catalog *services.CatalogService
	v       *validator.Validator
}

func NewCatalogHandler(c *services.CatalogService, v *validator.Validator) *CatalogHandler {
	return &CatalogHandler{catalog: c, v: v}
}

// Register mounts public (open) and admin (guarded) catalog routes.
func (h *CatalogHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	// Public
	api.GET("/menu", h.GetMenu)
	api.GET("/categories", h.ListCategories)
	api.GET("/products/:slug", h.GetProduct)

	// Admin
	admin := api.Group("/admin", jwtAuth, adminOnly)
	admin.GET("/categories", h.AdminListCategories)
	admin.POST("/categories", h.CreateCategory)
	admin.PUT("/categories/:id", h.UpdateCategory)
	admin.PATCH("/categories/:id", h.PatchCategory)
	admin.DELETE("/categories/:id", h.DeleteCategory)

	admin.GET("/products", h.AdminListProducts)
	admin.POST("/products", h.CreateProduct)
	admin.PUT("/products/:id", h.UpdateProduct)
	admin.PATCH("/products/:id/availability", h.SetAvailability)
	admin.DELETE("/products/:id", h.DeleteProduct)

	admin.POST("/cache/flush", h.FlushCache)
}

// locale extracts the requested locale from ?lang= or the Accept-Language
// header; empty means the service applies the default locale.
func locale(c echo.Context) string {
	if l := c.QueryParam("lang"); l != "" {
		return l
	}
	return c.Request().Header.Get("Accept-Language")
}

// --- Public ---

func (h *CatalogHandler) GetMenu(c echo.Context) error {
	menu, err := h.catalog.GetMenu(c.Request().Context(), locale(c))
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, menu)
}

func (h *CatalogHandler) ListCategories(c echo.Context) error {
	cats, err := h.catalog.ListCategories(c.Request().Context(), locale(c), true)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, cats)
}

func (h *CatalogHandler) GetProduct(c echo.Context) error {
	p, err := h.catalog.GetProductBySlug(c.Request().Context(), c.Param("slug"), locale(c))
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, p)
}

// --- Admin: list ---

func (h *CatalogHandler) AdminListCategories(c echo.Context) error {
	cats, err := h.catalog.AdminListCategories(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, cats)
}

func (h *CatalogHandler) AdminListProducts(c echo.Context) error {
	var categoryID uint64
	if raw := c.QueryParam("category_id"); raw != "" {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
			categoryID = id
		}
	}
	products, err := h.catalog.AdminListProducts(c.Request().Context(), categoryID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, products)
}

// PatchCategory handles partial updates (currently just is_active toggle).
func (h *CatalogHandler) PatchCategory(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var body struct {
		IsActive *bool `json:"is_active"`
	}
	if err := c.Bind(&body); err != nil {
		return response.BadRequest(c, "invalid request body")
	}
	if body.IsActive == nil {
		return response.BadRequest(c, "is_active is required")
	}
	cat, err := h.catalog.SetCategoryActive(c.Request().Context(), id, *body.IsActive)
	if err != nil {
		return mapCatalogError(c, err)
	}
	return response.OK(c, cat)
}

// --- Admin: categories ---

func (h *CatalogHandler) CreateCategory(c echo.Context) error {
	var in services.CategoryInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	cat, err := h.catalog.CreateCategory(c.Request().Context(), in)
	if err != nil {
		return mapCatalogError(c, err)
	}
	return response.Created(c, cat)
}

func (h *CatalogHandler) UpdateCategory(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.CategoryInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	cat, err := h.catalog.UpdateCategory(c.Request().Context(), id, in)
	if err != nil {
		return mapCatalogError(c, err)
	}
	return response.OK(c, cat)
}

func (h *CatalogHandler) DeleteCategory(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.catalog.DeleteCategory(c.Request().Context(), id); err != nil {
		return mapCatalogError(c, err)
	}
	return response.NoContent(c)
}

// --- Admin: products ---

func (h *CatalogHandler) CreateProduct(c echo.Context) error {
	var in services.ProductInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	p, err := h.catalog.CreateProduct(c.Request().Context(), in)
	if err != nil {
		return mapCatalogError(c, err)
	}
	return response.Created(c, p)
}

func (h *CatalogHandler) UpdateProduct(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.ProductInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	p, err := h.catalog.UpdateProduct(c.Request().Context(), id, in)
	if err != nil {
		return mapCatalogError(c, err)
	}
	return response.OK(c, p)
}

func (h *CatalogHandler) SetAvailability(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.AvailabilityInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	if err := h.catalog.SetProductAvailability(c.Request().Context(), id, in.IsAvailable); err != nil {
		return mapCatalogError(c, err)
	}
	return response.NoContent(c)
}

func (h *CatalogHandler) DeleteProduct(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.catalog.DeleteProduct(c.Request().Context(), id); err != nil {
		return mapCatalogError(c, err)
	}
	return response.NoContent(c)
}

func (h *CatalogHandler) FlushCache(c echo.Context) error {
	if err := h.catalog.FlushCache(c.Request().Context()); err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, echo.Map{"flushed": true})
}

// mapCatalogError handles catalog-specific conflicts on top of the shared map.
func mapCatalogError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, services.ErrSlugTaken):
		return response.Error(c, http.StatusConflict, response.CodeConflict, "slug already in use")
	case errors.Is(err, services.ErrCategoryNotEmpty):
		return response.Error(c, http.StatusConflict, response.CodeConflict, "category still has products")
	case errors.Is(err, services.ErrDefaultLocaleRequired):
		return response.Error(c, http.StatusUnprocessableEntity, response.CodeValidation, "a translation for the default locale is required")
	default:
		return mapServiceError(c, err)
	}
}

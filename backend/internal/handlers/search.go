package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// SearchHandler exposes public product search and an admin full-reindex.
type SearchHandler struct {
	search *services.SearchService
}

func NewSearchHandler(s *services.SearchService) *SearchHandler {
	return &SearchHandler{search: s}
}

func (h *SearchHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	api.GET("/search", h.Search)
	api.POST("/admin/search/reindex", h.Reindex, jwtAuth, adminOnly)
}

// Search handles GET /search?q=...&lang=...&limit=...
func (h *SearchHandler) Search(c echo.Context) error {
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	results, err := h.search.Search(c.Request().Context(), c.QueryParam("q"), locale(c), limit)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, results)
}

// Reindex rebuilds the whole Meilisearch index from MySQL.
func (h *SearchHandler) Reindex(c echo.Context) error {
	n, err := h.search.Reindex(c.Request().Context())
	if err != nil {
		// Reindex failures are almost always "Meilisearch is unreachable".
		return response.Error(c, http.StatusServiceUnavailable, response.CodeUnavailable,
			"search index is unavailable; ensure Meilisearch is running")
	}
	return response.OK(c, echo.Map{"indexed": n})
}

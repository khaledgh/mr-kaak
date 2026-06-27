package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/pkg/pagination"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// mediaSvc is the subset of MediaService used by the handler.
type mediaSvc interface {
	Upload(ctx context.Context, in services.UploadInput) (*services.MediaItem, error)
	List(ctx context.Context, p pagination.Params, q string) ([]services.MediaItem, int64, error)
	Delete(ctx context.Context, id uint64) error
}

// MediaHandler handles media upload, listing, and deletion.
type MediaHandler struct{ svc mediaSvc }

func NewMediaHandler(svc mediaSvc) *MediaHandler { return &MediaHandler{svc: svc} }

// Register mounts the three media endpoints under /admin/media, all guarded by
// JWT + admin-only middleware. The upload route overrides the global 1M body
// limit with 3M so the handler can do its own 2MB check and return a clean 422.
func (h *MediaHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	g := api.Group("/admin/media", jwtAuth, adminOnly)
	g.POST("", h.upload, emw.BodyLimit("3M"))
	g.GET("", h.list)
	g.DELETE("/:id", h.delete)
}

func (h *MediaHandler) upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file field is required")
	}

	src, err := file.Open()
	if err != nil {
		return response.BadRequest(c, "could not open uploaded file")
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return response.Internal(c)
	}

	var userID *uint64
	if uid, ok := middleware.UserIDFrom(c); ok {
		userID = &uid
	}

	item, err := h.svc.Upload(c.Request().Context(), services.UploadInput{
		Data:         data,
		OriginalName: file.Filename,
		UserID:       userID,
	})
	if err != nil {
		if errors.Is(err, services.ErrMediaTooLarge) {
			return response.Error(c, http.StatusUnprocessableEntity, response.CodeValidation,
				"file must not exceed 2 MB")
		}
		if errors.Is(err, services.ErrMediaInvalidType) {
			return response.Error(c, http.StatusUnprocessableEntity, response.CodeValidation,
				"only JPEG, PNG, and WebP images are allowed")
		}
		return mapServiceError(c, err)
	}

	return response.Created(c, item)
}

func (h *MediaHandler) list(c echo.Context) error {
	p := pagination.FromQuery(c)
	q := c.QueryParam("q")

	items, total, err := h.svc.List(c.Request().Context(), p, q)
	if err != nil {
		return mapServiceError(c, err)
	}

	return response.Paginated(c, http.StatusOK, items, pagination.NewMeta(p, total))
}

func (h *MediaHandler) delete(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return mapServiceError(c, err)
	}

	return response.NoContent(c)
}


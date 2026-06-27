package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/pagination"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// UserHandler exposes the caller's address book and admin user management.
type UserHandler struct {
	users *services.UserService
	v     *validator.Validator
}

func NewUserHandler(u *services.UserService, v *validator.Validator) *UserHandler {
	return &UserHandler{users: u, v: v}
}

// Register mounts routes. jwtAuth guards all of them; adminOnly additionally
// guards the /admin/users/* management routes.
func (h *UserHandler) Register(api *echo.Group, jwtAuth, adminOnly echo.MiddlewareFunc) {
	// Self-service address book (any authenticated user).
	addr := api.Group("/me/addresses", jwtAuth)
	addr.GET("", h.ListAddresses)
	addr.POST("", h.AddAddress)
	addr.PUT("/:id", h.UpdateAddress)
	addr.DELETE("/:id", h.DeleteAddress)
	addr.POST("/:id/default", h.SetDefaultAddress)

	// Admin user management.
	admin := api.Group("/admin/users", jwtAuth, adminOnly)
	admin.GET("", h.ListUsers)
	admin.GET("/:id", h.GetUser)
	admin.PATCH("/:id/role", h.ChangeRole)
	admin.POST("/:id/suspend", h.Suspend)
	admin.POST("/:id/activate", h.Activate)
}

// --- Addresses (self) ---

func (h *UserHandler) ListAddresses(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	addrs, err := h.users.ListAddresses(c.Request().Context(), uid)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, addrs)
}

func (h *UserHandler) AddAddress(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	var in services.AddressInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	a, err := h.users.AddAddress(c.Request().Context(), uid, in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, a)
}

func (h *UserHandler) UpdateAddress(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.AddressInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	a, err := h.users.UpdateAddress(c.Request().Context(), uid, id, in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, a)
}

func (h *UserHandler) DeleteAddress(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.users.DeleteAddress(c.Request().Context(), uid, id); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

func (h *UserHandler) SetDefaultAddress(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.users.SetDefaultAddress(c.Request().Context(), uid, id); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

// --- Admin user management ---

func (h *UserHandler) ListUsers(c echo.Context) error {
	p := pagination.FromQuery(c)
	opts := repository.UserListOptions{
		Role:   c.QueryParam("role"),
		Search: c.QueryParam("q"),
		Limit:  p.Limit(),
		Offset: p.Offset(),
	}
	users, total, err := h.users.ListUsers(c.Request().Context(), opts)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Paginated(c, 200, users, pagination.NewMeta(p, total))
}

func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	u, err := h.users.GetByID(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, u)
}

func (h *UserHandler) ChangeRole(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	var in services.ChangeRoleInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	u, err := h.users.ChangeRole(c.Request().Context(), id, models.Role(in.Role))
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, u)
}

func (h *UserHandler) Suspend(c echo.Context) error {
	return h.setStatus(c, models.UserSuspended)
}

func (h *UserHandler) Activate(c echo.Context) error {
	return h.setStatus(c, models.UserActive)
}

func (h *UserHandler) setStatus(c echo.Context, status models.UserStatus) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.users.SetStatus(c.Request().Context(), id, status); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

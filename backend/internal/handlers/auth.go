package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/auth"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// AuthHandler exposes registration, login, refresh, logout, and profile.
type AuthHandler struct {
	auth  *services.AuthService
	users *services.UserService
	v     *validator.Validator
}

func NewAuthHandler(a *services.AuthService, u *services.UserService, v *validator.Validator) *AuthHandler {
	return &AuthHandler{auth: a, users: u, v: v}
}

// Register mounts public auth routes on api (v1). jwtAuth guards the
// authenticated profile endpoints.
func (h *AuthHandler) Register(api *echo.Group, jwtAuth echo.MiddlewareFunc) {
	g := api.Group("/auth")
	g.POST("/register", h.RegisterUser)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)

	g.POST("/logout", h.Logout, jwtAuth)
	g.GET("/me", h.Me, jwtAuth)
	g.PATCH("/me", h.UpdateMe, jwtAuth)
}

// authResponse is the success body for register/login/refresh.
type authResponse struct {
	User   *models.User   `json:"user"`
	Tokens auth.TokenPair `json:"tokens"`
}

func (h *AuthHandler) RegisterUser(c echo.Context) error {
	var in services.RegisterInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	res, err := h.auth.Register(c.Request().Context(), in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.Created(c, authResponse{User: res.User, Tokens: res.Tokens})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var in services.LoginInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	res, err := h.auth.Login(c.Request().Context(), in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, authResponse{User: res.User, Tokens: res.Tokens})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var in services.RefreshInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	res, err := h.auth.Refresh(c.Request().Context(), in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, authResponse{User: res.User, Tokens: res.Tokens})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	if err := h.auth.Logout(c.Request().Context(), uid); err != nil {
		return mapServiceError(c, err)
	}
	return response.NoContent(c)
}

func (h *AuthHandler) Me(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	u, err := h.users.GetByID(c.Request().Context(), uid)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, u)
}

func (h *AuthHandler) UpdateMe(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	var in services.UpdateProfileInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	u, err := h.users.UpdateProfile(c.Request().Context(), uid, in)
	if err != nil {
		return mapServiceError(c, err)
	}
	return response.OK(c, u)
}

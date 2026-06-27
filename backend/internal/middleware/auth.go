package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/auth"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// Context keys for authenticated identity. Unexported so only the accessors
// below can read/write them.
const (
	ctxUserID = "auth_user_id"
	ctxRole   = "auth_role"
)

// Auth returns middleware that requires a valid access token. On success it
// stores the user id and role in the Echo context. It performs no DB lookup —
// the short-lived access token is trusted from its signature (stateless auth).
func Auth(jwt *auth.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw, err := bearerToken(c)
			if err != nil {
				return response.Unauthorized(c, "missing or malformed Authorization header")
			}
			claims, err := jwt.ParseAccess(raw)
			if err != nil {
				return response.Unauthorized(c, "invalid or expired token")
			}
			uid, err := claims.UserID()
			if err != nil {
				return response.Unauthorized(c, "invalid token subject")
			}
			c.Set(ctxUserID, uid)
			c.Set(ctxRole, claims.Role)
			return next(c)
		}
	}
}

// RequireRole returns middleware that allows only the given roles. Must be
// chained after Auth. super_admin always passes.
func RequireRole(allowed ...models.Role) echo.MiddlewareFunc {
	allow := make(map[models.Role]struct{}, len(allowed))
	for _, r := range allowed {
		allow[r] = struct{}{}
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := RoleFrom(c)
			if !ok {
				return response.Unauthorized(c, "authentication required")
			}
			if role == models.RoleSuperAdmin {
				return next(c)
			}
			if _, ok := allow[role]; !ok {
				return response.Forbidden(c, "insufficient permissions")
			}
			return next(c)
		}
	}
}

// RequireStaff allows any back-office role (kitchen/staff/admin/super_admin).
func RequireStaff() echo.MiddlewareFunc {
	return RequireRole(models.RoleKitchen, models.RoleStaff, models.RoleAdmin)
}

// UserIDFrom returns the authenticated user id, or false if unauthenticated.
func UserIDFrom(c echo.Context) (uint64, bool) {
	v, ok := c.Get(ctxUserID).(uint64)
	return v, ok
}

// RoleFrom returns the authenticated user's role, or false if unauthenticated.
func RoleFrom(c echo.Context) (models.Role, bool) {
	v, ok := c.Get(ctxRole).(models.Role)
	return v, ok
}

func bearerToken(c echo.Context) (string, error) {
	h := c.Request().Header.Get(echo.HeaderAuthorization)
	if h == "" {
		// SSE fallback: EventSource can't set headers, so allow the access token
		// as a query param for stream endpoints (token in URL is acceptable here).
		if q := c.QueryParam("access_token"); q != "" {
			return q, nil
		}
		return "", echo.ErrUnauthorized
	}
	const prefix = "Bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", echo.ErrUnauthorized
	}
	return strings.TrimSpace(h[len(prefix):]), nil
}

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

// errHandled signals that an error response has already been written to the
// client. Handlers return it to stop processing; the central error handler
// sees the committed response and does nothing further. This is essential:
// the response helpers return c.JSON's result (nil on success), so without a
// sentinel a failed bind/validate would return nil and the handler would keep
// running and persist invalid data.
var errHandled = errors.New("response already written")

// bindValidate binds the JSON body into dst and runs validation. On failure it
// writes the appropriate error envelope and returns errHandled so the caller
// returns immediately without further processing.
func bindValidate(c echo.Context, v *validator.Validator, dst any) error {
	if err := c.Bind(dst); err != nil {
		_ = response.BadRequest(c, "invalid request body")
		return errHandled
	}
	if err := v.Struct(dst); err != nil {
		var ve *validator.ValidationError
		if errors.As(err, &ve) {
			_ = response.Error(c, http.StatusUnprocessableEntity,
				response.CodeValidation, "validation failed", ve.Fields)
			return errHandled
		}
		_ = response.BadRequest(c, "invalid request")
		return errHandled
	}
	return nil
}

// mapServiceError converts a domain error from the services layer into the
// standard error envelope. Centralizing this keeps handlers free of HTTP
// status bookkeeping.
func mapServiceError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, services.ErrNotFound):
		return response.NotFound(c, "resource not found")
	case errors.Is(err, services.ErrEmailTaken):
		return response.Error(c, http.StatusConflict, response.CodeConflict, "email already registered")
	case errors.Is(err, services.ErrInvalidCredentials):
		return response.Unauthorized(c, "invalid email or password")
	case errors.Is(err, services.ErrAccountSuspended):
		return response.Forbidden(c, "account suspended")
	case errors.Is(err, services.ErrInvalidToken):
		return response.Unauthorized(c, "invalid or expired token")
	case errors.Is(err, services.ErrForbidden):
		return response.Forbidden(c, "forbidden")
	default:
		// Unknown error: log via the central handler and return a generic 500.
		return err
	}
}

// idParam parses a uint64 path parameter (e.g. /users/:id).
func idParam(c echo.Context, name string) (uint64, error) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		_ = response.BadRequest(c, "invalid "+name)
		return 0, errHandled
	}
	return id, nil
}

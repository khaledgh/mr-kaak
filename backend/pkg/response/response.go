// Package response defines the single JSON envelope every endpoint returns,
// plus a stable set of machine-readable error codes. A consistent envelope
// lets the frontend handle success and failure uniformly.
package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Envelope is the canonical response body.
//
//	success:  { "data": {...}, "meta": {...}? }
//	failure:  { "error": { "code": "...", "message": "...", "details": {...}? } }
type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Meta  any        `json:"meta,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

// ErrorBody is the machine-readable error payload.
type ErrorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Code is a stable, client-facing error identifier (do not renumber/rename
// without coordinating with the frontend).
type Code string

const (
	CodeBadRequest     Code = "bad_request"
	CodeValidation     Code = "validation_error"
	CodeUnauthorized   Code = "unauthorized"
	CodeForbidden      Code = "forbidden"
	CodeNotFound       Code = "not_found"
	CodeConflict       Code = "conflict"
	CodeRateLimited    Code = "rate_limited"
	CodeUnprocessable  Code = "unprocessable"
	CodePaymentFailed  Code = "payment_failed"
	CodeInternal       Code = "internal_error"
	CodeUnavailable    Code = "service_unavailable"
	CodeIdempotencyHit Code = "idempotency_conflict"
)

// JSON writes a success envelope.
func JSON(c echo.Context, status int, data any) error {
	return c.JSON(status, Envelope{Data: data})
}

// Paginated writes a success envelope with pagination metadata in meta.
func Paginated(c echo.Context, status int, data any, meta any) error {
	return c.JSON(status, Envelope{Data: data, Meta: meta})
}

// Error writes a failure envelope.
func Error(c echo.Context, status int, code Code, message string, details ...any) error {
	body := &ErrorBody{Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details[0]
	}
	return c.JSON(status, Envelope{Error: body})
}

// Helpers for the common cases — keep handlers terse and consistent.

func OK(c echo.Context, data any) error      { return JSON(c, http.StatusOK, data) }
func Created(c echo.Context, data any) error { return JSON(c, http.StatusCreated, data) }
func NoContent(c echo.Context) error         { return c.NoContent(http.StatusNoContent) }

func NotFound(c echo.Context, msg string) error {
	return Error(c, http.StatusNotFound, CodeNotFound, msg)
}

func BadRequest(c echo.Context, msg string, details ...any) error {
	return Error(c, http.StatusBadRequest, CodeBadRequest, msg, details...)
}

func Unauthorized(c echo.Context, msg string) error {
	return Error(c, http.StatusUnauthorized, CodeUnauthorized, msg)
}

func Forbidden(c echo.Context, msg string) error {
	return Error(c, http.StatusForbidden, CodeForbidden, msg)
}

func Internal(c echo.Context) error {
	// Never leak internal detail to clients; the real cause is logged server-side.
	return Error(c, http.StatusInternalServerError, CodeInternal, "an unexpected error occurred")
}

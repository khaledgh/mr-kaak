package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mrkaak/restaurant-api/internal/middleware"
	"github.com/mrkaak/restaurant-api/internal/services"
	"github.com/mrkaak/restaurant-api/internal/validator"
	"github.com/mrkaak/restaurant-api/pkg/response"
)

// PaymentHandler exposes available methods, card-payment finalization, the
// Square webhook, and admin refunds.
type PaymentHandler struct {
	payments *services.PaymentService
	v        *validator.Validator
}

func NewPaymentHandler(p *services.PaymentService, v *validator.Validator) *PaymentHandler {
	return &PaymentHandler{payments: p, v: v}
}

func (h *PaymentHandler) Register(api *echo.Group, jwtAuth, staffOnly echo.MiddlewareFunc) {
	api.GET("/payment/methods", h.Methods, jwtAuth)
	api.POST("/payments/square", h.ConfirmCard, jwtAuth)
	// Webhook is public but signature-verified inside the service.
	api.POST("/payments/square/webhook", h.SquareWebhook)
	// Admin refund.
	api.POST("/admin/orders/:id/refund", h.Refund, jwtAuth, staffOnly)
}

func (h *PaymentHandler) Methods(c echo.Context) error {
	return response.OK(c, echo.Map{"methods": h.payments.AvailableMethods(c.Request().Context())})
}

type confirmCardInput struct {
	OrderCode   string `json:"order_code" validate:"required"`
	SourceToken string `json:"source_token" validate:"required"`
}

func (h *PaymentHandler) ConfirmCard(c echo.Context) error {
	uid, _ := middleware.UserIDFrom(c)
	var in confirmCardInput
	if err := bindValidate(c, h.v, &in); err != nil {
		return err
	}
	idem := c.Request().Header.Get("Idempotency-Key")
	if idem == "" {
		idem = "card-" + in.OrderCode
	}
	o, err := h.payments.ConfirmCardPayment(c.Request().Context(), uid, in.OrderCode, in.SourceToken, idem)
	if err != nil {
		return mapPaymentError(c, err)
	}
	return response.OK(c, o)
}

func (h *PaymentHandler) SquareWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.BadRequest(c, "cannot read body")
	}
	signature := c.Request().Header.Get("x-square-hmacsha256-signature")
	// The signed URL is the full public notification URL Square called.
	requestURL := c.Scheme() + "://" + c.Request().Host + c.Request().RequestURI

	if err := h.payments.HandleSquareWebhook(c.Request().Context(), body, signature, requestURL); err != nil {
		if errors.Is(err, services.ErrWebhookInvalid) {
			return response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid webhook signature")
		}
		return mapServiceError(c, err)
	}
	return response.OK(c, echo.Map{"received": true})
}

func (h *PaymentHandler) Refund(c echo.Context) error {
	id, err := idParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.payments.Refund(c.Request().Context(), id); err != nil {
		return mapPaymentError(c, err)
	}
	return response.OK(c, echo.Map{"refunded": true})
}

func mapPaymentError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, services.ErrPaymentMethodDisabled):
		return response.Error(c, http.StatusUnprocessableEntity, response.CodeUnprocessable, "payment method not available")
	case errors.Is(err, services.ErrPaymentFailed):
		return response.Error(c, http.StatusPaymentRequired, response.CodePaymentFailed, "payment failed")
	case errors.Is(err, services.ErrNothingToRefund):
		return response.Error(c, http.StatusUnprocessableEntity, response.CodeUnprocessable, "no captured payment to refund")
	case errors.Is(err, services.ErrRefundFailed):
		return response.Error(c, http.StatusBadGateway, response.CodePaymentFailed, "refund failed")
	default:
		return mapServiceError(c, err)
	}
}

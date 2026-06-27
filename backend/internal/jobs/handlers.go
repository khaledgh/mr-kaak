package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/internal/services"
)

// Handlers wires task processors to the services they need.
type Handlers struct {
	search *services.SearchService
	push   *services.PushService
	meta   *services.MetaService
	orders *repository.OrderRepo
	log    *slog.Logger
}

func NewHandlers(search *services.SearchService, push *services.PushService, meta *services.MetaService, orders *repository.OrderRepo, log *slog.Logger) *Handlers {
	return &Handlers{search: search, push: push, meta: meta, orders: orders, log: log}
}

// Register mounts every task handler on the Asynq mux.
func (h *Handlers) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeSearchIndex, h.handleSearchIndex)
	mux.HandleFunc(TypeSearchDelete, h.handleSearchDelete)
	mux.HandleFunc(TypeNotifyOrder, h.handleNotifyOrder)
	mux.HandleFunc(TypeMetaPurchase, h.handleMetaPurchase)
}

func (h *Handlers) handleMetaPurchase(ctx context.Context, t *asynq.Task) error {
	var p MetaPurchasePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w: %w", err, asynq.SkipRetry)
	}
	return h.meta.SendPurchase(ctx, p.OrderID)
}

func (h *Handlers) handleSearchIndex(ctx context.Context, t *asynq.Task) error {
	var p SearchIndexPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w: %w", err, asynq.SkipRetry)
	}
	return h.search.IndexProduct(ctx, p.ProductID)
}

func (h *Handlers) handleSearchDelete(ctx context.Context, t *asynq.Task) error {
	var p SearchDeletePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w: %w", err, asynq.SkipRetry)
	}
	return h.search.DeleteProductFromIndex(ctx, p.ProductID)
}

// handleNotifyOrder sends a Web Push notification to the order's owner on a
// status change (a no-op if push isn't configured or there are no subscriptions).
func (h *Handlers) handleNotifyOrder(ctx context.Context, t *asynq.Task) error {
	var p NotifyOrderPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal: %w: %w", err, asynq.SkipRetry)
	}
	o, err := h.orders.FindByID(ctx, p.OrderID)
	if err != nil {
		return err
	}
	return h.push.SendToUser(ctx, o.UserID, services.Notification{
		Title: "Order " + o.Code,
		Body:  statusMessage(p.Status),
		URL:   "/orders/" + o.Code,
	})
}

// statusMessage maps an order status to a customer-facing notification body.
func statusMessage(status string) string {
	switch status {
	case "confirmed":
		return "Your order is confirmed 🎉"
	case "preparing":
		return "We're preparing your order 👨‍🍳"
	case "ready":
		return "Your order is ready ✅"
	case "out_for_delivery":
		return "Your order is out for delivery 🚗"
	case "delivered":
		return "Your order has been delivered. Enjoy! 😋"
	case "cancelled":
		return "Your order was cancelled."
	default:
		return "Your order status changed to " + status
	}
}

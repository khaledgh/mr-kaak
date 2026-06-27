package jobs

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/pkg/logger"
)

// Enqueuer submits tasks to the Asynq queue (Redis). Enqueue failures are
// best-effort and logged, never fatal — so the HTTP request that triggered the
// side effect still succeeds even if Redis is temporarily down.
type Enqueuer struct {
	client *asynq.Client
}

func NewEnqueuer(cfg config.Redis) *Enqueuer {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB,
	})
	return &Enqueuer{client: client}
}

func (e *Enqueuer) Close() error { return e.client.Close() }

// IndexProduct enqueues a reindex for a product (best-effort).
func (e *Enqueuer) IndexProduct(ctx context.Context, productID uint64) {
	t, err := NewSearchIndexTask(productID)
	if err == nil {
		e.enqueue(ctx, t)
	}
}

// DeleteProduct enqueues a de-index for a product (best-effort).
func (e *Enqueuer) DeleteProduct(ctx context.Context, productID uint64) {
	t, err := NewSearchDeleteTask(productID)
	if err == nil {
		e.enqueue(ctx, t)
	}
}

// NotifyOrderStatus enqueues an order-status notification (best-effort).
func (e *Enqueuer) NotifyOrderStatus(ctx context.Context, orderID uint64, status string) {
	t, err := NewNotifyOrderTask(orderID, status)
	if err == nil {
		e.enqueue(ctx, t)
	}
}

// TrackPurchase enqueues a Meta Conversions API Purchase event (best-effort).
func (e *Enqueuer) TrackPurchase(ctx context.Context, orderID uint64) {
	t, err := NewMetaPurchaseTask(orderID)
	if err == nil {
		e.enqueue(ctx, t)
	}
}

func (e *Enqueuer) enqueue(ctx context.Context, t *asynq.Task) {
	if _, err := e.client.EnqueueContext(ctx, t); err != nil {
		logger.FromContext(ctx).Warn("enqueue task failed (continuing)",
			slog.String("type", t.Type()), slog.Any("err", err))
	}
}

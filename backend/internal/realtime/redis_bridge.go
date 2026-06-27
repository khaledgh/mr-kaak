package realtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// channel is the Redis pub/sub channel carrying order events across replicas.
const channel = "orders:events"

// Publisher publishes an event both locally and (when Redis is up) to other
// replicas. The Hub alone covers a single instance; the Redis fan-out covers N.
type Publisher struct {
	hub *Hub
	rdb *redis.Client
}

func NewPublisher(hub *Hub, rdb *redis.Client) *Publisher {
	return &Publisher{hub: hub, rdb: rdb}
}

// PublishOrderEvent is the convenience form used by the order service.
func (p *Publisher) PublishOrderEvent(ctx context.Context, orderCode, status, message string) {
	p.Publish(ctx, Event{OrderCode: orderCode, Status: status, Message: message})
}

// Publish delivers locally and best-effort to Redis (ignored if Redis is down).
func (p *Publisher) Publish(ctx context.Context, e Event) {
	p.hub.Publish(e)
	if p.rdb != nil {
		if buf, err := json.Marshal(e); err == nil {
			_ = p.rdb.Publish(ctx, channel, buf).Err()
		}
	}
}

// RunBridge subscribes to the Redis channel and republishes remote events into
// the local hub, so a status change on any replica reaches clients on this one.
// Returns when ctx is cancelled. Safe to skip entirely if Redis is unavailable.
func RunBridge(ctx context.Context, hub *Hub, rdb *redis.Client, log *slog.Logger) {
	if rdb == nil {
		return
	}
	sub := rdb.Subscribe(ctx, channel)
	defer sub.Close()
	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var e Event
			if err := json.Unmarshal([]byte(msg.Payload), &e); err != nil {
				log.Warn("realtime bridge: bad payload", slog.Any("err", err))
				continue
			}
			hub.Publish(e)
		}
	}
}

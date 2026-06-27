package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/mrkaak/restaurant-api/internal/config"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
	"github.com/mrkaak/restaurant-api/pkg/logger"
)

// PushService stores Web Push subscriptions and sends notifications (plan §8).
// If VAPID keys are not configured, sends are no-ops (push simply disabled).
type PushService struct {
	subs  *repository.PushRepo
	vapid config.VAPID
}

func NewPushService(subs *repository.PushRepo, vapid config.VAPID) *PushService {
	return &PushService{subs: subs, vapid: vapid}
}

// Enabled reports whether push is configured.
func (s *PushService) Enabled() bool {
	return s.vapid.PublicKey != "" && s.vapid.PrivateKey != ""
}

// PublicKey is exposed to the frontend so it can subscribe.
func (s *PushService) PublicKey() string { return s.vapid.PublicKey }

// Subscribe stores (or refreshes) a browser push subscription.
func (s *PushService) Subscribe(ctx context.Context, userID uint64, in PushSubscribeInput) error {
	return s.subs.Upsert(ctx, &models.PushSubscription{
		UserID: userID, Endpoint: in.Endpoint, P256dh: in.Keys.P256dh,
		Auth: in.Keys.Auth, UserAgent: in.UserAgent,
	})
}

// Unsubscribe removes a subscription by endpoint.
func (s *PushService) Unsubscribe(ctx context.Context, userID uint64, endpoint string) error {
	return s.subs.DeleteByEndpoint(ctx, userID, endpoint)
}

// Notification is the payload delivered to the service worker.
type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"`
}

// SendToUser delivers a notification to all of a user's subscriptions, pruning
// endpoints the push service reports as gone (404/410).
func (s *PushService) SendToUser(ctx context.Context, userID uint64, n Notification) error {
	if !s.Enabled() {
		return nil
	}
	subs, err := s.subs.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	payload, _ := json.Marshal(n)
	log := logger.FromContext(ctx)

	for _, sub := range subs {
		resp, err := webpush.SendNotification(payload, &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
		}, &webpush.Options{
			Subscriber:      s.vapid.Subject,
			VAPIDPublicKey:  s.vapid.PublicKey,
			VAPIDPrivateKey: s.vapid.PrivateKey,
			TTL:             60,
		})
		if err != nil {
			log.Warn("push send failed", slog.Uint64("sub", sub.ID), slog.Any("err", err))
			continue
		}
		// A 404/410 means the subscription is dead — prune it.
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			_ = s.subs.DeleteByID(ctx, sub.ID)
		}
		resp.Body.Close()
	}
	return nil
}

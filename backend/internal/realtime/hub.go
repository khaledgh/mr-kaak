// Package realtime delivers live order-status updates to clients over SSE.
// It uses an in-process hub (works for a single instance immediately) and an
// optional Redis pub/sub bridge so updates fan out across API replicas
// (plan §9). With Redis down, same-instance delivery still works.
package realtime

import (
	"sync"
)

// Event is a status update pushed to subscribers of an order.
type Event struct {
	OrderCode string `json:"order_code"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

// Hub is an in-process fan-out: many SSE connections subscribe per order code.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan Event]struct{} // orderCode -> set of subscriber channels
}

func NewHub() *Hub {
	return &Hub{subs: make(map[string]map[chan Event]struct{})}
}

// Subscribe returns a channel of events for an order code and an unsubscribe
// func the caller must defer.
func (h *Hub) Subscribe(orderCode string) (<-chan Event, func()) {
	ch := make(chan Event, 8)
	h.mu.Lock()
	set := h.subs[orderCode]
	if set == nil {
		set = make(map[chan Event]struct{})
		h.subs[orderCode] = set
	}
	set[ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		if set := h.subs[orderCode]; set != nil {
			delete(set, ch)
			if len(set) == 0 {
				delete(h.subs, orderCode)
			}
		}
		h.mu.Unlock()
		close(ch)
	}
}

// Publish delivers an event to all local subscribers of the order. Non-blocking:
// a slow/full subscriber is skipped rather than stalling the publisher.
func (h *Hub) Publish(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[e.OrderCode] {
		select {
		case ch <- e:
		default: // drop for a slow consumer; it will re-fetch state on reconnect
		}
	}
}

// Package jobs defines Asynq task types and a thin enqueuer. Tasks offload
// non-critical/heavy work off the HTTP request path (plan §3.3d): search
// reindexing, notifications, emails, payment-webhook processing.
package jobs

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Queue names with implied priority (configured in the worker).
const (
	QueueCritical = "critical" // payments
	QueueDefault  = "default"  // notifications
	QueueLow      = "low"      // reindex, analytics
)

// Task type identifiers.
const (
	TypeSearchIndex  = "search:index"
	TypeSearchDelete = "search:delete"
	TypeNotifyOrder  = "notify:order_status"
	TypeMetaPurchase = "meta:purchase"
)

// SearchIndexPayload reindexes one product.
type SearchIndexPayload struct {
	ProductID uint64 `json:"product_id"`
}

// SearchDeletePayload removes one product from the index.
type SearchDeletePayload struct {
	ProductID uint64 `json:"product_id"`
}

// NotifyOrderPayload pushes an order-status notification.
type NotifyOrderPayload struct {
	OrderID uint64 `json:"order_id"`
	Status  string `json:"status"`
}

func newTask(typ string, payload any, opts ...asynq.Option) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(typ, b, opts...), nil
}

// NewSearchIndexTask builds a low-priority reindex task.
func NewSearchIndexTask(productID uint64) (*asynq.Task, error) {
	return newTask(TypeSearchIndex, SearchIndexPayload{ProductID: productID}, asynq.Queue(QueueLow))
}

// NewSearchDeleteTask builds a low-priority de-index task.
func NewSearchDeleteTask(productID uint64) (*asynq.Task, error) {
	return newTask(TypeSearchDelete, SearchDeletePayload{ProductID: productID}, asynq.Queue(QueueLow))
}

// NewNotifyOrderTask builds a default-priority notification task.
func NewNotifyOrderTask(orderID uint64, status string) (*asynq.Task, error) {
	return newTask(TypeNotifyOrder, NotifyOrderPayload{OrderID: orderID, Status: status}, asynq.Queue(QueueDefault))
}

// MetaPurchasePayload sends a server-side Purchase conversion event.
type MetaPurchasePayload struct {
	OrderID uint64 `json:"order_id"`
}

// NewMetaPurchaseTask builds a low-priority Meta Conversions API task.
func NewMetaPurchaseTask(orderID uint64) (*asynq.Task, error) {
	return newTask(TypeMetaPurchase, MetaPurchasePayload{OrderID: orderID}, asynq.Queue(QueueLow))
}

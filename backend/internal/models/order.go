package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// OrderStatus values and the allowed transition graph (plan §4).
type OrderStatus string

const (
	StatusPendingPayment OrderStatus = "pending_payment"
	StatusPaid           OrderStatus = "paid"
	StatusConfirmed      OrderStatus = "confirmed"
	StatusPreparing      OrderStatus = "preparing"
	StatusReady          OrderStatus = "ready"
	StatusOutForDelivery OrderStatus = "out_for_delivery"
	StatusDelivered      OrderStatus = "delivered"
	StatusCancelled      OrderStatus = "cancelled"
	StatusRefunded       OrderStatus = "refunded"
)

// allowedTransitions maps a status to the statuses it may advance to.
var allowedTransitions = map[OrderStatus][]OrderStatus{
	StatusPendingPayment: {StatusPaid, StatusConfirmed, StatusCancelled},
	StatusPaid:           {StatusConfirmed, StatusCancelled, StatusRefunded},
	StatusConfirmed:      {StatusPreparing, StatusCancelled, StatusRefunded},
	StatusPreparing:      {StatusReady, StatusCancelled},
	StatusReady:          {StatusOutForDelivery, StatusDelivered, StatusCancelled},
	StatusOutForDelivery: {StatusDelivered, StatusCancelled},
	StatusDelivered:      {StatusRefunded},
	StatusCancelled:      {},
	StatusRefunded:       {},
}

// CanTransitionTo reports whether moving from s to next is allowed.
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	for _, allowed := range allowedTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

// PaymentStatus values.
type PaymentStatus string

const (
	PayPending    PaymentStatus = "pending"
	PayPaid       PaymentStatus = "paid"
	PayCODPending PaymentStatus = "cod_pending"
	PayFailed     PaymentStatus = "failed"
	PayRefunded   PaymentStatus = "refunded"
)

type FulfillmentType string

const (
	FulfillmentDelivery FulfillmentType = "delivery"
	FulfillmentPickup   FulfillmentType = "pickup"
)

// AddressSnapshot is the delivery address copied onto the order at checkout so
// later edits to the user's address book never alter historical orders.
type AddressSnapshot struct {
	Label        string   `json:"label,omitempty"`
	Line1        string   `json:"line1"`
	Line2        string   `json:"line2,omitempty"`
	City         string   `json:"city"`
	ProvinceCode string   `json:"province_code"`
	PostalCode   string   `json:"postal_code"`
	CountryCode  string   `json:"country_code"`
	Lat          *float64 `json:"lat,omitempty"`
	Lng          *float64 `json:"lng,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	Notes        string   `json:"notes,omitempty"`
}

func (a *AddressSnapshot) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *AddressSnapshot) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return errors.New("address snapshot: unsupported scan type")
	}
}

// ModifierSnapshot records a chosen modifier on an order line.
type ModifierSnapshot struct {
	ModifierID      uint64 `json:"modifier_id"`
	Label           string `json:"label"`
	PriceDeltaCents int64  `json:"price_delta_cents"`
}

// ModifierSnapshots is a JSON array of chosen modifiers on an order item.
type ModifierSnapshots []ModifierSnapshot

func (m ModifierSnapshots) Value() (driver.Value, error) {
	if m == nil {
		return "[]", nil
	}
	return json.Marshal(m)
}

func (m *ModifierSnapshots) Scan(src any) error {
	if src == nil {
		*m = ModifierSnapshots{}
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return errors.New("modifier snapshots: unsupported scan type")
	}
}

// Order is a placed order with snapshotted pricing.
type Order struct {
	Base
	UserID           uint64           `gorm:"index;not null" json:"user_id"`
	Code             string           `gorm:"size:20;uniqueIndex;not null" json:"code"`
	Status           OrderStatus      `gorm:"size:24;not null;default:pending_payment;index" json:"status"`
	FulfillmentType  FulfillmentType  `gorm:"type:enum('delivery','pickup');not null;default:delivery" json:"fulfillment_type"`
	SubtotalCents    int64            `gorm:"not null;default:0" json:"subtotal_cents"`
	DiscountCents    int64            `gorm:"not null;default:0" json:"discount_cents"`
	DeliveryFeeCents int64            `gorm:"not null;default:0" json:"delivery_fee_cents"`
	TaxCents         int64            `gorm:"not null;default:0" json:"tax_cents"`
	TotalCents       int64            `gorm:"not null;default:0" json:"total_cents"`
	Currency         string           `gorm:"size:3;not null;default:CAD" json:"currency"`
	PaymentMethod    string           `gorm:"size:20;not null;default:cod" json:"payment_method"`
	PaymentStatus    PaymentStatus    `gorm:"size:24;not null;default:pending" json:"payment_status"`
	CouponID         *uint64          `json:"coupon_id,omitempty"`
	AddressSnapshot  *AddressSnapshot `gorm:"column:address_snapshot_json;type:json" json:"address_snapshot,omitempty"`
	ScheduledFor     *time.Time       `json:"scheduled_for,omitempty"`
	Notes            string           `gorm:"size:500" json:"notes,omitempty"`
	IdempotencyKey   *string          `gorm:"column:idempotency_key;size:80" json:"-"`

	Items   []OrderItem          `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	History []OrderStatusHistory `gorm:"foreignKey:OrderID" json:"history,omitempty"`
}

// OrderItem is one line of an order with snapshotted name and price.
type OrderItem struct {
	ID             uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID        uint64            `gorm:"index;not null" json:"order_id"`
	ProductID      uint64            `gorm:"not null" json:"product_id"`
	VariantID      *uint64           `json:"variant_id,omitempty"`
	NameSnapshot   string            `gorm:"size:200;not null" json:"name_snapshot"`
	Qty            int               `gorm:"not null;default:1" json:"qty"`
	WeightGrams    *int              `json:"weight_grams,omitempty"`
	UnitPriceCents int64             `gorm:"not null;default:0" json:"unit_price_cents"`
	Modifiers      ModifierSnapshots `gorm:"column:modifiers_json;type:json" json:"modifiers"`
	LineTotalCents int64             `gorm:"not null;default:0" json:"line_total_cents"`
	CreatedAt      time.Time         `json:"created_at"`
}

// TableName pins the singular table name (GORM would otherwise pluralize it to
// "order_status_histories").
func (OrderStatusHistory) TableName() string { return "order_status_history" }

// OrderStatusHistory records every status transition (audit trail, plan §17).
type OrderStatusHistory struct {
	ID         uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID    uint64      `gorm:"index;not null" json:"order_id"`
	FromStatus OrderStatus `gorm:"size:24" json:"from_status,omitempty"`
	ToStatus   OrderStatus `gorm:"size:24;not null" json:"to_status"`
	ActorID    *uint64     `json:"actor_id,omitempty"`
	Note       string      `gorm:"size:255" json:"note,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

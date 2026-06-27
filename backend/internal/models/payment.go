package models

import (
	"encoding/json"
	"time"
)

// PaymentTxnKind distinguishes a charge from a refund.
type PaymentTxnKind string

const (
	TxnCharge PaymentTxnKind = "charge"
	TxnRefund PaymentTxnKind = "refund"
)

// PaymentTxnStatus is the lifecycle state of a transaction.
type PaymentTxnStatus string

const (
	TxnPending   PaymentTxnStatus = "pending"
	TxnSucceeded PaymentTxnStatus = "succeeded"
	TxnFailed    PaymentTxnStatus = "failed"
	TxnRefunded  PaymentTxnStatus = "refunded"
)

// PaymentTransaction records one provider interaction for an order.
type PaymentTransaction struct {
	ID          uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     uint64           `gorm:"index;not null" json:"order_id"`
	Provider    string           `gorm:"size:20;not null" json:"provider"`
	ProviderRef string           `gorm:"size:120" json:"provider_ref,omitempty"`
	Kind        PaymentTxnKind   `gorm:"size:20;not null;default:charge" json:"kind"`
	AmountCents int64            `gorm:"not null;default:0" json:"amount_cents"`
	Currency    string           `gorm:"size:3;not null;default:CAD" json:"currency"`
	Status      PaymentTxnStatus `gorm:"size:24;not null" json:"status"`
	RawPayload  json.RawMessage  `gorm:"column:raw_payload_json;type:json" json:"-"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

package models

import "time"

type PaymentProvider string

const PaymentProviderEsewa PaymentProvider = "esewa"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment is one payment attempt against an order. It is its own row rather
// than a few columns on Order because an order and its attempt to pay for it
// are different lifecycles: the order is what was bought, the payment is
// whether (and how) it got paid for, and keeping them separate is what lets
// a failed attempt be recorded without it looking like the order itself
// failed to exist.
type Payment struct {
	ID       uint            `gorm:"primaryKey" json:"id"`
	OrderID  uint            `gorm:"not null;index" json:"order_id"`
	Order    Order           `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"-"`
	Provider PaymentProvider `gorm:"not null" json:"provider"`
	// TransactionUUID is this app's own idempotency key for the attempt --
	// what eSewa calls transaction_uuid. Generated once, before the redirect,
	// and never reused, so any signature or status-check response can always
	// be traced back to exactly one payment row.
	TransactionUUID string        `gorm:"not null;uniqueIndex" json:"transaction_uuid"`
	Amount          float64       `gorm:"not null" json:"amount"`
	Status          PaymentStatus `gorm:"not null;default:pending" json:"status"`
	// ProviderRefID is eSewa's own reference for a completed transaction --
	// transaction_code on the success callback, ref_id on the independent
	// status-check call, the same value under two names depending which
	// response it came from. Nil until the payment actually completes.
	ProviderRefID *string `json:"provider_ref_id"`
	// RawResponse is eSewa's raw status-check response, kept whole for
	// support and debugging rather than trusting the handful of fields
	// parsed out of it to tell the entire story.
	RawResponse RawJSON   `gorm:"type:jsonb" json:"raw_response"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

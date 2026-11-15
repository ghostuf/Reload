package models

import "time"

// OrderStatus tracks the lifecycle of an order. The mock checkout path still
// jumps straight to "completed" -- there is nothing to wait on -- but the
// eSewa path spends real time in "pending" between redirect and verified
// payment, and can land on "failed" or "cancelled" instead of ever
// completing.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	UserID     uint        `gorm:"not null;index" json:"user_id"`
	User       User        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	TotalPrice float64     `gorm:"not null" json:"total_price"`
	Status     OrderStatus `gorm:"not null;default:completed" json:"status"`
	Items      []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	// Payment is the (at most one, in this app) payment attempt against the
	// order. Nil for mock-checkout orders, which never create one.
	Payment   *Payment  `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

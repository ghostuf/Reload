package models

import "time"

// OrderStatus tracks the lifecycle of an order. This is a mock checkout flow
// with no real payment processor, so every order lands as "completed".
type OrderStatus string

const OrderStatusCompleted OrderStatus = "completed"

type Order struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	UserID     uint        `gorm:"not null;index" json:"user_id"`
	User       User        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	TotalPrice float64     `gorm:"not null" json:"total_price"`
	Status     OrderStatus `gorm:"not null;default:completed" json:"status"`
	Items      []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

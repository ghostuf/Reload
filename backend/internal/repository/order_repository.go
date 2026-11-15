package repository

import (
	"fmt"
	"time"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// ListByUser returns a user's orders, newest first, with items preloaded so
// a summary (item count, total) can be built without a second query per
// order.
func (r *OrderRepository) ListByUser(userID uint) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Items").
		Preload("Payment").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("listing orders: %w", err)
	}
	return orders, nil
}

// Create inserts the order and its items together. Called against a
// transaction-scoped repository during checkout, so the order and its items
// land atomically with the library grants and cart clear.
func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

// GetByID returns a single order with its items and each item's game
// preloaded, scoped to userID. It returns gorm.ErrRecordNotFound both when
// the order does not exist and when it belongs to someone else, so callers
// can't distinguish the two -- which is what keeps this from leaking whether
// an order id exists on another account.
//
// The game preload is Unscoped so a receipt still shows what was bought even
// after an admin soft-deletes that game -- the whole point of a soft delete
// over a real one.
func (r *OrderRepository) GetByID(userID, orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items.Game", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}).
		Preload("Payment").
		Where("user_id = ?", userID).
		First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// MarkStatus sets an order's status directly, bypassing Save so it never
// touches total_price or any other column.
func (r *OrderRepository) MarkStatus(id uint, status models.OrderStatus) error {
	if err := r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("updating order status: %w", err)
	}
	return nil
}

// ExpireStalePending marks a user's own pending orders -- and their pending
// payments -- as cancelled once they have sat unpaid for longer than
// maxAge. There is no background job in this app; this runs as a lazy check
// wherever a user's orders are read, which is often enough that a stale
// session never has to wait for anything else to notice it.
func (r *OrderRepository) ExpireStalePending(userID uint, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	return r.db.Transaction(func(tx *gorm.DB) error {
		var staleIDs []uint
		err := tx.Model(&models.Order{}).
			Where("user_id = ? AND status = ? AND created_at < ?", userID, models.OrderStatusPending, cutoff).
			Pluck("id", &staleIDs).Error
		if err != nil {
			return fmt.Errorf("finding stale pending orders: %w", err)
		}
		if len(staleIDs) == 0 {
			return nil
		}

		// Payments first, while the orders they belong to are still
		// findable as "pending" -- flipping the order first would leave
		// nothing distinguishing "just expired" from "pending from any
		// other order" if this were ever run concurrently.
		err = tx.Model(&models.Payment{}).
			Where("order_id IN ? AND status = ?", staleIDs, models.PaymentStatusPending).
			Update("status", models.PaymentStatusCancelled).Error
		if err != nil {
			return fmt.Errorf("expiring stale payments: %w", err)
		}

		if err := tx.Model(&models.Order{}).Where("id IN ?", staleIDs).
			Update("status", models.OrderStatusCancelled).Error; err != nil {
			return fmt.Errorf("expiring stale orders: %w", err)
		}
		return nil
	})
}

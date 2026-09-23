package repository

import (
	"fmt"

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
func (r *OrderRepository) GetByID(userID, orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items.Game").
		Where("user_id = ?", userID).
		First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

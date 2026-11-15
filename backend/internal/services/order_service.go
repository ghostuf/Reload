package services

import (
	"fmt"
	"time"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
)

type OrderStore interface {
	ListByUser(userID uint) ([]models.Order, error)
	GetByID(userID, orderID uint) (*models.Order, error)
	ExpireStalePending(userID uint, maxAge time.Duration) error
}

var _ OrderStore = (*repository.OrderRepository)(nil)

type OrderService struct {
	orders OrderStore
}

func NewOrderService(orders OrderStore) *OrderService {
	return &OrderService{orders: orders}
}

// ListOrders first lazily expires any of the user's own pending orders that
// have sat unpaid past staleOrderAge -- there is no background job in this
// app, so a read is the only reliable place to notice one.
func (s *OrderService) ListOrders(userID uint) ([]models.Order, error) {
	if err := s.orders.ExpireStalePending(userID, staleOrderAge); err != nil {
		return nil, fmt.Errorf("expiring stale orders: %w", err)
	}

	orders, err := s.orders.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("loading orders: %w", err)
	}
	return orders, nil
}

// GetOrder returns gorm.ErrRecordNotFound when the order does not exist or
// does not belong to userID; OrderRepository.GetByID scopes the lookup to
// userID so the two cases are indistinguishable, which is what keeps this
// from leaking another user's order ids.
func (s *OrderService) GetOrder(userID, orderID uint) (*models.Order, error) {
	if err := s.orders.ExpireStalePending(userID, staleOrderAge); err != nil {
		return nil, fmt.Errorf("expiring stale orders: %w", err)
	}

	order, err := s.orders.GetByID(userID, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}

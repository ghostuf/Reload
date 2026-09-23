package services

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
)

type OrderStore interface {
	ListByUser(userID uint) ([]models.Order, error)
	GetByID(userID, orderID uint) (*models.Order, error)
}

var _ OrderStore = (*repository.OrderRepository)(nil)

type OrderService struct {
	orders OrderStore
}

func NewOrderService(orders OrderStore) *OrderService {
	return &OrderService{orders: orders}
}

func (s *OrderService) ListOrders(userID uint) ([]models.Order, error) {
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
	order, err := s.orders.GetByID(userID, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}

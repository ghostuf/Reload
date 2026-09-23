package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

// ErrCartEmpty is returned when there is nothing in the user's cart to check
// out.
var ErrCartEmpty = errors.New("cart is empty")

// ErrNothingToPurchase is returned when every game in the cart is already in
// the user's library, so there is nothing left to charge for.
var ErrNothingToPurchase = errors.New("all games in your cart are already in your library")

// CheckoutService holds the raw *gorm.DB, rather than a repository
// interface, because checkout has to run its reads and writes across cart,
// order, and library tables inside one transaction. Each step below
// constructs its repository against the transaction handle so every write
// commits -- or rolls back -- together.
type CheckoutService struct {
	db *gorm.DB
}

func NewCheckoutService(db *gorm.DB) *CheckoutService {
	return &CheckoutService{db: db}
}

// Checkout turns a user's cart into an order: it snapshots each game's
// current price into an order item, grants library access for each one, and
// empties the cart -- all inside one transaction, so a failure partway
// through (say, the library grant fails) cannot leave an order created with
// an uncleared cart or a charge without a library grant.
//
// Games already in the user's library are skipped rather than re-purchased,
// so re-running checkout on a stale cart (e.g. a duplicate tab) never
// double-charges or double-owns.
func (s *CheckoutService) Checkout(userID uint) (*models.Order, error) {
	var orderID uint

	err := s.db.Transaction(func(tx *gorm.DB) error {
		cartRepo := repository.NewCartRepository(tx)
		libraryRepo := repository.NewLibraryRepository(tx)
		orderRepo := repository.NewOrderRepository(tx)

		cartItems, err := cartRepo.ListByUser(userID)
		if err != nil {
			return fmt.Errorf("loading cart: %w", err)
		}
		if len(cartItems) == 0 {
			return ErrCartEmpty
		}

		owned, err := libraryRepo.OwnedGameIDs(userID)
		if err != nil {
			return fmt.Errorf("loading library: %w", err)
		}

		var (
			orderItems []models.OrderItem
			total      float64
		)
		for _, item := range cartItems {
			if owned[item.GameID] {
				continue
			}
			price := effectivePrice(item.Game)
			orderItems = append(orderItems, models.OrderItem{
				GameID:          item.GameID,
				PriceAtPurchase: price,
			})
			total += price
		}

		if len(orderItems) == 0 {
			return ErrNothingToPurchase
		}

		order := models.Order{
			UserID:     userID,
			TotalPrice: total,
			Status:     models.OrderStatusCompleted,
			Items:      orderItems,
		}
		if err := orderRepo.Create(&order); err != nil {
			return fmt.Errorf("creating order: %w", err)
		}

		acquiredAt := time.Now()
		libraryItems := make([]models.LibraryItem, len(orderItems))
		for i, item := range orderItems {
			libraryItems[i] = models.LibraryItem{
				UserID:     userID,
				GameID:     item.GameID,
				AcquiredAt: acquiredAt,
			}
		}
		if err := libraryRepo.CreateMany(libraryItems); err != nil {
			return fmt.Errorf("granting library access: %w", err)
		}

		if err := cartRepo.ClearByUser(userID); err != nil {
			return fmt.Errorf("clearing cart: %w", err)
		}

		orderID = order.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload outside the transaction with each item's game preloaded, so the
	// response carries full game details rather than just IDs.
	order, err := repository.NewOrderRepository(s.db).GetByID(userID, orderID)
	if err != nil {
		return nil, fmt.Errorf("loading created order: %w", err)
	}
	return order, nil
}

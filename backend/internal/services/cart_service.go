package services

import (
	"errors"
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

var ErrAlreadyInCart = errors.New("game is already in cart")

type CartStore interface {
	ListByUser(userID uint) ([]models.CartItem, error)
	Create(item *models.CartItem) error
	Delete(userID, gameID uint) error
}

var (
	_ CartStore = (*repository.CartRepository)(nil)
	_ GameStore = (*repository.GameRepository)(nil)
)

type CartService struct {
	cart  CartStore
	games GameStore
}

func NewCartService(cart CartStore, games GameStore) *CartService {
	return &CartService{cart: cart, games: games}
}

type CartResult struct {
	Items []models.CartItem
	Total float64
}

func (s *CartService) GetCart(userID uint) (CartResult, error) {
	items, err := s.cart.ListByUser(userID)
	if err != nil {
		return CartResult{}, fmt.Errorf("loading cart: %w", err)
	}

	var total float64
	for _, item := range items {
		total += effectivePrice(item.Game)
	}

	return CartResult{Items: items, Total: total}, nil
}

// AddToCart validates the game exists before writing, so a bad game_id fails
// with ErrGameNotFound rather than a DB foreign-key error.
func (s *CartService) AddToCart(userID, gameID uint) (*models.CartItem, error) {
	game, err := s.games.GetByID(gameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("looking up game: %w", err)
	}

	item := &models.CartItem{UserID: userID, GameID: gameID}
	if err := s.cart.Create(item); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyInCart
		}
		return nil, fmt.Errorf("adding to cart: %w", err)
	}

	item.Game = *game
	return item, nil
}

// RemoveFromCart reports gorm.ErrRecordNotFound when the game was not in the
// user's cart.
func (s *CartService) RemoveFromCart(userID, gameID uint) error {
	return s.cart.Delete(userID, gameID)
}

// effectivePrice is what a customer actually pays for a game: the discount
// price when one is set, otherwise the list price.
func effectivePrice(game models.Game) float64 {
	if game.DiscountPrice != nil {
		return *game.DiscountPrice
	}
	return game.Price
}

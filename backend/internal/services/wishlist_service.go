package services

import (
	"errors"
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

var ErrAlreadyInWishlist = errors.New("game is already in wishlist")

type WishlistStore interface {
	ListByUser(userID uint) ([]models.WishlistItem, error)
	Create(item *models.WishlistItem) error
	Delete(userID, gameID uint) error
}

var _ WishlistStore = (*repository.WishlistRepository)(nil)

type WishlistService struct {
	wishlist WishlistStore
	games    GameStore
}

func NewWishlistService(wishlist WishlistStore, games GameStore) *WishlistService {
	return &WishlistService{wishlist: wishlist, games: games}
}

func (s *WishlistService) GetWishlist(userID uint) ([]models.WishlistItem, error) {
	items, err := s.wishlist.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("loading wishlist: %w", err)
	}
	return items, nil
}

// AddToWishlist validates the game exists before writing, so a bad game_id
// fails with ErrGameNotFound rather than a DB foreign-key error.
func (s *WishlistService) AddToWishlist(userID, gameID uint) (*models.WishlistItem, error) {
	game, err := s.games.GetByID(gameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("looking up game: %w", err)
	}

	item := &models.WishlistItem{UserID: userID, GameID: gameID}
	if err := s.wishlist.Create(item); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyInWishlist
		}
		return nil, fmt.Errorf("adding to wishlist: %w", err)
	}

	item.Game = *game
	return item, nil
}

// RemoveFromWishlist reports gorm.ErrRecordNotFound when the game was not in
// the user's wishlist.
func (s *WishlistService) RemoveFromWishlist(userID, gameID uint) error {
	return s.wishlist.Delete(userID, gameID)
}

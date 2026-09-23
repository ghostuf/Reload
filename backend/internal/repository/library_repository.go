package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type LibraryRepository struct {
	db *gorm.DB
}

func NewLibraryRepository(db *gorm.DB) *LibraryRepository {
	return &LibraryRepository{db: db}
}

// ListByUser returns a user's owned games with full game details preloaded,
// most recently acquired first.
func (r *LibraryRepository) ListByUser(userID uint) ([]models.LibraryItem, error) {
	var items []models.LibraryItem
	err := r.db.Preload("Game").
		Preload("Game.Genres").
		Where("user_id = ?", userID).
		Order("acquired_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("listing library: %w", err)
	}
	return items, nil
}

// OwnedGameIDs returns the set of game IDs a user already owns. Checkout
// uses this to skip games already in the library rather than double-charge.
func (r *LibraryRepository) OwnedGameIDs(userID uint) (map[uint]bool, error) {
	var gameIDs []uint
	err := r.db.Model(&models.LibraryItem{}).
		Where("user_id = ?", userID).
		Pluck("game_id", &gameIDs).Error
	if err != nil {
		return nil, fmt.Errorf("loading owned games: %w", err)
	}

	owned := make(map[uint]bool, len(gameIDs))
	for _, id := range gameIDs {
		owned[id] = true
	}
	return owned, nil
}

// CreateMany grants library access for the given items in one insert.
func (r *LibraryRepository) CreateMany(items []models.LibraryItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LibraryRepository struct {
	db *gorm.DB
}

func NewLibraryRepository(db *gorm.DB) *LibraryRepository {
	return &LibraryRepository{db: db}
}

// ListByUser returns a user's owned games with full game details preloaded,
// most recently acquired first. The Game preload is Unscoped so a game an
// admin has since soft-deleted still shows up as owned rather than vanishing
// from a library entry that is otherwise still valid.
func (r *LibraryRepository) ListByUser(userID uint) ([]models.LibraryItem, error) {
	var items []models.LibraryItem
	err := r.db.Preload("Game", func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	}).
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

// CreateManyIgnoringConflicts grants library access for the given items,
// silently skipping any that violate the (user_id, game_id) unique
// constraint rather than failing the whole call. eSewa verification uses
// this instead of CreateMany: a payment confirmed after the user somehow
// already owns one of the games (a race with another session, say) must
// still grant the rest rather than roll the entire transaction back over
// one already-owned item.
func (r *LibraryRepository) CreateManyIgnoringConflicts(items []models.LibraryItem) error {
	if len(items) == 0 {
		return nil
	}
	err := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&items).Error
	if err != nil {
		return fmt.Errorf("granting library access: %w", err)
	}
	return nil
}

// HasGame reports whether a user owns a specific game. Reviews check this
// before accepting one, so it is a single count rather than loading the
// user's whole library to look for one game_id.
func (r *LibraryRepository) HasGame(userID, gameID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.LibraryItem{}).
		Where("user_id = ? AND game_id = ?", userID, gameID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("checking library ownership: %w", err)
	}
	return count > 0, nil
}

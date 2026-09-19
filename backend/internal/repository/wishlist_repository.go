package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type WishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) *WishlistRepository {
	return &WishlistRepository{db: db}
}

// ListByUser returns a user's wishlist items with game details preloaded,
// oldest addition first.
func (r *WishlistRepository) ListByUser(userID uint) ([]models.WishlistItem, error) {
	var items []models.WishlistItem
	err := r.db.Preload("Game").
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("listing wishlist items: %w", err)
	}
	return items, nil
}

// Create inserts a wishlist item. A duplicate (user_id, game_id) surfaces as
// gorm.ErrDuplicatedKey.
func (r *WishlistRepository) Create(item *models.WishlistItem) error {
	return r.db.Create(item).Error
}

// Delete removes a user's wishlist item for a game. It returns
// gorm.ErrRecordNotFound when there was nothing to remove.
func (r *WishlistRepository) Delete(userID, gameID uint) error {
	result := r.db.Where("user_id = ? AND game_id = ?", userID, gameID).
		Delete(&models.WishlistItem{})
	if result.Error != nil {
		return fmt.Errorf("removing wishlist item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

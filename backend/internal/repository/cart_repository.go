package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// ListByUser returns a user's cart items with game details preloaded,
// oldest addition first.
func (r *CartRepository) ListByUser(userID uint) ([]models.CartItem, error) {
	var items []models.CartItem
	err := r.db.Preload("Game").
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("listing cart items: %w", err)
	}
	return items, nil
}

// Create inserts a cart item. A duplicate (user_id, game_id) surfaces as
// gorm.ErrDuplicatedKey.
func (r *CartRepository) Create(item *models.CartItem) error {
	return r.db.Create(item).Error
}

// Delete removes a user's cart item for a game. It returns
// gorm.ErrRecordNotFound when there was nothing to remove.
func (r *CartRepository) Delete(userID, gameID uint) error {
	result := r.db.Where("user_id = ? AND game_id = ?", userID, gameID).
		Delete(&models.CartItem{})
	if result.Error != nil {
		return fmt.Errorf("removing cart item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

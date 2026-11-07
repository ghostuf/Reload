package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type ReviewFilter struct {
	GameID   uint
	Page     int
	PageSize int
}

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// ListByGame returns a game's reviews, newest first, with each reviewer's
// user record preloaded so the display name can be shown without a query per
// review.
func (r *ReviewRepository) ListByGame(filter ReviewFilter) ([]models.Review, int64, error) {
	base := r.db.Model(&models.Review{}).Where("game_id = ?", filter.GameID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting reviews: %w", err)
	}

	offset := (filter.Page - 1) * filter.PageSize

	var reviews []models.Review
	err := base.Preload("User").
		Order("created_at DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, fmt.Errorf("listing reviews: %w", err)
	}

	return reviews, total, nil
}

// Create inserts a review. A duplicate (user_id, game_id) surfaces as
// gorm.ErrDuplicatedKey, the same unique index that limits a user to one
// review per game.
func (r *ReviewRepository) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

// GetByID returns a single review, or gorm.ErrRecordNotFound.
func (r *ReviewRepository) GetByID(id uint) (*models.Review, error) {
	var review models.Review
	if err := r.db.First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

// Update saves a review's rating and body. Select lists exactly those two
// columns so a body cleared back to empty is still written -- a plain
// struct-based Updates call would otherwise skip it as a Go zero value.
func (r *ReviewRepository) Update(review *models.Review) error {
	result := r.db.Model(&models.Review{}).
		Where("id = ?", review.ID).
		Select("rating", "body").
		Updates(review)
	if result.Error != nil {
		return fmt.Errorf("updating review: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete removes a review. It returns gorm.ErrRecordNotFound when there was
// nothing to remove.
func (r *ReviewRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Review{}, id)
	if result.Error != nil {
		return fmt.Errorf("deleting review: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Aggregate returns a game's average rating and review count, computed in
// the database rather than by loading every review. Avg comes back nil when
// the game has no reviews, since SQL's AVG() over zero rows is NULL.
func (r *ReviewRepository) Aggregate(gameID uint) (*float64, int64, error) {
	var result struct {
		Avg   *float64
		Count int64
	}

	err := r.db.Model(&models.Review{}).
		Select("AVG(rating) AS avg, COUNT(*) AS count").
		Where("game_id = ?", gameID).
		Scan(&result).Error
	if err != nil {
		return nil, 0, fmt.Errorf("aggregating reviews: %w", err)
	}

	return result.Avg, result.Count, nil
}

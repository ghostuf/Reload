package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type GameFilter struct {
	Search   string
	Genre    string
	Sort     string
	Page     int
	PageSize int
}

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) List(filter GameFilter) ([]models.Game, int64, error) {
	base := r.db.Model(&models.Game{})

	if filter.Search != "" {
		base = base.Where("title ILIKE ?", "%"+filter.Search+"%")
	}

	if filter.Genre != "" {
		base = base.Joins("JOIN game_genres ON game_genres.game_id = games.id").
			Joins("JOIN genres ON genres.id = game_genres.genre_id").
			Where("genres.name ILIKE ?", filter.Genre)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting games: %w", err)
	}

	query := base.Preload("Genres").Order(sortClause(filter.Sort))

	offset := (filter.Page - 1) * filter.PageSize
	query = query.Limit(filter.PageSize).Offset(offset)

	var games []models.Game
	if err := query.Find(&games).Error; err != nil {
		return nil, 0, fmt.Errorf("listing games: %w", err)
	}

	return games, total, nil
}

func (r *GameRepository) GetByID(id uint) (*models.Game, error) {
	var game models.Game

	err := r.db.
		Preload("Genres").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("game_images.sort_order ASC")
		}).
		First(&game, id).Error
	if err != nil {
		return nil, err
	}

	return &game, nil
}

func sortClause(sort string) string {
	switch sort {
	case "price_asc":
		return "price ASC"
	case "price_desc":
		return "price DESC"
	case "title_asc":
		return "title ASC"
	case "title_desc":
		return "title DESC"
	case "release_date":
		return "release_date DESC"
	default:
		return "created_at DESC"
	}
}

package repository

import (
	"fmt"
	"time"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/gorm"
)

type GameFilter struct {
	Search   string
	Genre    string
	OnSale   bool
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
		// search_vector is a generated tsvector over title and description
		// (see enableGameSearch); plainto_tsquery tolerates arbitrary user
		// input rather than requiring tsquery's own operator syntax.
		base = base.Where("search_vector @@ plainto_tsquery('english', ?)", filter.Search)
	}

	if filter.Genre != "" {
		base = base.Joins("JOIN game_genres ON game_genres.game_id = games.id").
			Joins("JOIN genres ON genres.id = game_genres.genre_id").
			Where("genres.name ILIKE ?", filter.Genre)
	}

	if filter.OnSale {
		base = base.Where("discount_price IS NOT NULL AND (discount_ends_at IS NULL OR discount_ends_at > ?)", time.Now())
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

// Create inserts a game together with its genres and gallery images. Genres
// must already carry real primary keys (see ResolveGenres) so this only ever
// writes join rows for them, never duplicate genre records.
func (r *GameRepository) Create(game *models.Game) error {
	return r.db.Create(game).Error
}

// Update replaces a game's fields, genres, and gallery images. It is a full
// PUT-style replacement rather than a partial patch: genres and images not
// present in game are removed, and scalar fields -- including clearing a
// discount back to nil -- are written even when the new value is a Go zero
// value, which a plain struct-based Updates call would otherwise skip.
func (r *GameRepository) Update(game *models.Game) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Game{}).
			Where("id = ?", game.ID).
			Select(
				"title", "slug", "description", "developer", "publisher",
				"release_date", "price", "discount_price", "discount_ends_at",
				"cover_image_url",
			).
			Updates(game)
		if result.Error != nil {
			return fmt.Errorf("updating game: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Model(game).Association("Genres").Replace(game.Genres); err != nil {
			return fmt.Errorf("updating genres: %w", err)
		}

		if err := tx.Where("game_id = ?", game.ID).Delete(&models.GameImage{}).Error; err != nil {
			return fmt.Errorf("clearing images: %w", err)
		}
		if len(game.Images) > 0 {
			images := make([]models.GameImage, len(game.Images))
			for i, img := range game.Images {
				images[i] = models.GameImage{GameID: game.ID, ImageURL: img.ImageURL, SortOrder: i}
			}
			if err := tx.Create(&images).Error; err != nil {
				return fmt.Errorf("creating images: %w", err)
			}
		}

		return nil
	})
}

// Delete soft-deletes a game: Game has a DeletedAt field, so this is an
// UPDATE setting it rather than a real DELETE, leaving every foreign key an
// existing order or library entry holds on the game intact. It reports
// gorm.ErrRecordNotFound when there was nothing to delete.
func (r *GameRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Game{}, id)
	if result.Error != nil {
		return fmt.Errorf("deleting game: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ResolveGenres maps genre names onto existing Genre rows, creating any that
// don't exist yet -- the same FirstOrCreate-by-name approach the seed script
// uses, so admin-created games share genre rows with seeded ones instead of
// spawning duplicates that only differ by casing of an otherwise-same name.
func (r *GameRepository) ResolveGenres(names []string) ([]models.Genre, error) {
	genres := make([]models.Genre, 0, len(names))
	for _, name := range names {
		genre := models.Genre{Name: name}
		if err := r.db.Where("name = ?", name).FirstOrCreate(&genre).Error; err != nil {
			return nil, fmt.Errorf("resolving genre %q: %w", name, err)
		}
		genres = append(genres, genre)
	}
	return genres, nil
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

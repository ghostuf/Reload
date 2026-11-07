package repository

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		// Surfaces unique-constraint violations as gorm.ErrDuplicatedKey rather
		// than a driver-specific error string.
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Genre{},
		&models.Game{},
		&models.GameImage{},
		&models.User{},
		&models.RefreshToken{},
		&models.CartItem{},
		&models.WishlistItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.LibraryItem{},
		&models.Review{},
	); err != nil {
		return fmt.Errorf("running auto-migration: %w", err)
	}

	if err := enableGameSearch(db); err != nil {
		return err
	}

	return nil
}

// enableGameSearch adds a generated tsvector column over a game's title and
// description, weighted so a title match ranks above a description match,
// plus a GIN index over it. A generated column is not something GORM's
// struct tags can express, so this runs once as raw SQL rather than through
// AutoMigrate; both statements are idempotent so re-running it on every
// startup is harmless.
func enableGameSearch(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE games ADD COLUMN IF NOT EXISTS search_vector tsvector
			GENERATED ALWAYS AS (
				setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
				setweight(to_tsvector('english', coalesce(description, '')), 'B')
			) STORED`,
		`CREATE INDEX IF NOT EXISTS idx_games_search_vector ON games USING GIN (search_vector)`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("enabling game search index: %w", err)
		}
	}

	return nil
}

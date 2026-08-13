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
	); err != nil {
		return fmt.Errorf("running auto-migration: %w", err)
	}
	return nil
}

package models

import (
	"time"
)

type Game struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	Title         string      `gorm:"not null" json:"title"`
	Slug          string      `gorm:"uniqueIndex;not null" json:"slug"`
	Description   string      `json:"description"`
	Developer     string      `json:"developer"`
	Publisher     string      `json:"publisher"`
	ReleaseDate   time.Time   `json:"release_date"`
	Price         float64     `gorm:"not null" json:"price"`
	DiscountPrice *float64    `json:"discount_price"`
	CoverImageURL string      `json:"cover_image_url"`
	Genres        []Genre     `gorm:"many2many:game_genres;" json:"genres,omitempty"`
	Images        []GameImage `gorm:"foreignKey:GameID" json:"images,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

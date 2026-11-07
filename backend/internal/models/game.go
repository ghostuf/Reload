package models

import (
	"time"

	"gorm.io/gorm"
)

type Game struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Title         string    `gorm:"not null" json:"title"`
	Slug          string    `gorm:"uniqueIndex;not null" json:"slug"`
	Description   string    `json:"description"`
	Developer     string    `json:"developer"`
	Publisher     string    `json:"publisher"`
	ReleaseDate   time.Time `json:"release_date"`
	Price         float64   `gorm:"not null" json:"price"`
	DiscountPrice *float64  `json:"discount_price"`
	// DiscountEndsAt is when a discount stops applying. A nil value means the
	// discount (if any) never expires on its own -- it is only turned off by
	// clearing DiscountPrice.
	DiscountEndsAt *time.Time  `json:"discount_ends_at"`
	CoverImageURL  string      `json:"cover_image_url"`
	Genres         []Genre     `gorm:"many2many:game_genres;" json:"genres,omitempty"`
	Images         []GameImage `gorm:"foreignKey:GameID" json:"images,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	// DeletedAt is a soft delete: removing a game hides it from the store
	// without breaking the foreign keys that existing orders and library
	// entries hold on it.
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

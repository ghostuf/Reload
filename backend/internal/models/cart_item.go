package models

import "time"

type CartItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_cart_items_user_game" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	GameID    uint      `gorm:"not null;uniqueIndex:idx_cart_items_user_game" json:"game_id"`
	Game      Game      `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"game"`
	CreatedAt time.Time `json:"created_at"`
}

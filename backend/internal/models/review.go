package models

import "time"

// Review is one user's rating and write-up of a game. The unique index on
// (user_id, game_id) is what makes "one review per game" a database
// guarantee rather than just an application rule.
type Review struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_reviews_user_game" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	GameID    uint      `gorm:"not null;uniqueIndex:idx_reviews_user_game" json:"game_id"`
	Game      Game      `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"-"`
	Rating    int       `gorm:"not null" json:"rating"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

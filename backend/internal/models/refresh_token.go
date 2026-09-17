package models

import (
	"time"
)

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Revoked   bool      `gorm:"not null;default:false" json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

package models

type GameImage struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	GameID    uint   `gorm:"not null;index" json:"game_id"`
	ImageURL  string `gorm:"not null" json:"image_url"`
	SortOrder int    `json:"sort_order"`
}

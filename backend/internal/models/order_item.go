package models

// OrderItem snapshots what a game cost at the moment of purchase. Price is
// stored here rather than read live off Game, so a later price change never
// rewrites the history of what a customer actually paid.
type OrderItem struct {
	ID              uint    `gorm:"primaryKey" json:"id"`
	OrderID         uint    `gorm:"not null;index" json:"order_id"`
	GameID          uint    `gorm:"not null" json:"game_id"`
	Game            Game    `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"game"`
	PriceAtPurchase float64 `gorm:"not null" json:"price_at_purchase"`
}

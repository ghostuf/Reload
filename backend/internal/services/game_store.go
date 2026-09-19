package services

import (
	"errors"

	"github.com/ghostuf/reload/internal/models"
)

// ErrGameNotFound is returned when an operation references a game_id that
// does not exist. Cart and wishlist both check this before writing, so a bad
// game_id surfaces as a 404 instead of a DB foreign-key error.
var ErrGameNotFound = errors.New("game not found")

// GameStore is the read slice of game data the cart and wishlist services
// need to validate a game_id before adding it.
type GameStore interface {
	GetByID(id uint) (*models.Game, error)
}

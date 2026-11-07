package services

import (
	"time"

	"github.com/ghostuf/reload/internal/models"
)

// effectivePrice is what a customer actually pays for a game: the discount
// price when one is set and still active, otherwise the list price. Every
// price a customer sees or pays -- game list and detail responses, cart
// totals, and checkout's price_at_purchase -- is computed here and nowhere
// else, so a rule about discounts only has to be correct in one place.
func effectivePrice(game models.Game) float64 {
	if game.DiscountPrice != nil && discountActive(game) {
		return *game.DiscountPrice
	}
	return game.Price
}

// discountActive reports whether a set discount_price is still on: it never
// expires on its own when discount_ends_at is nil, and otherwise it is active
// until that moment passes.
func discountActive(game models.Game) bool {
	return game.DiscountEndsAt == nil || game.DiscountEndsAt.After(time.Now())
}

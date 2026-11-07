package services

import (
	"testing"
	"time"

	"github.com/ghostuf/reload/internal/models"
)

func priceOf(f float64) *float64 { return &f }

func TestEffectivePrice(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	tests := map[string]struct {
		game models.Game
		want float64
	}{
		"no discount": {
			game: models.Game{Price: 59.99},
			want: 59.99,
		},
		"discount with no end date never expires": {
			game: models.Game{Price: 59.99, DiscountPrice: priceOf(39.99)},
			want: 39.99,
		},
		"discount ending in the future is active": {
			game: models.Game{Price: 59.99, DiscountPrice: priceOf(39.99), DiscountEndsAt: &future},
			want: 39.99,
		},
		"discount that already ended falls back to list price": {
			game: models.Game{Price: 59.99, DiscountPrice: priceOf(39.99), DiscountEndsAt: &past},
			want: 59.99,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := effectivePrice(tt.game); got != tt.want {
				t.Errorf("effectivePrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

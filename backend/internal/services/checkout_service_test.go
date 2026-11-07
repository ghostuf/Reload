package services

import (
	"testing"
	"time"

	"github.com/ghostuf/reload/internal/models"
)

// TestBuildOrderItemsAppliesActiveAndExpiredDiscounts covers the pricing rule
// that decides what checkout actually charges: an active discount wins, an
// expired one falls back to list price, and a game already in the library is
// skipped rather than charged for twice.
func TestBuildOrderItemsAppliesActiveAndExpiredDiscounts(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	cartItems := []models.CartItem{
		{
			GameID: 1,
			Game:   models.Game{Price: 59.99, DiscountPrice: priceOf(39.99), DiscountEndsAt: &future},
		},
		{
			GameID: 2,
			Game:   models.Game{Price: 24.99, DiscountPrice: priceOf(14.99), DiscountEndsAt: &past},
		},
		{
			GameID: 3,
			Game:   models.Game{Price: 19.99},
		},
		{
			// Already owned -- must be skipped entirely, not charged at list
			// price or any other price.
			GameID: 4,
			Game:   models.Game{Price: 9.99},
		},
	}
	owned := map[uint]bool{4: true}

	orderItems, total := buildOrderItems(cartItems, owned)

	if len(orderItems) != 3 {
		t.Fatalf("len(orderItems) = %d, want 3", len(orderItems))
	}

	prices := make(map[uint]float64, len(orderItems))
	for _, item := range orderItems {
		prices[item.GameID] = item.PriceAtPurchase
		if item.GameID == 4 {
			t.Error("an owned game was included in the order")
		}
	}

	if got, want := prices[1], 39.99; got != want {
		t.Errorf("active discount priced at %v, want %v", got, want)
	}
	if got, want := prices[2], 24.99; got != want {
		t.Errorf("expired discount priced at %v, want list price %v", got, want)
	}
	if got, want := prices[3], 19.99; got != want {
		t.Errorf("undiscounted game priced at %v, want %v", got, want)
	}

	wantTotal := 39.99 + 24.99 + 19.99
	if total != wantTotal {
		t.Errorf("total = %v, want %v", total, wantTotal)
	}
}

// TestBuildOrderItemsAllOwnedYieldsNothingToPurchase mirrors the caller's
// ErrNothingToPurchase check: a cart made entirely of already-owned games
// must build zero order items.
func TestBuildOrderItemsAllOwnedYieldsNothingToPurchase(t *testing.T) {
	cartItems := []models.CartItem{
		{GameID: 1, Game: models.Game{Price: 59.99}},
	}
	owned := map[uint]bool{1: true}

	orderItems, total := buildOrderItems(cartItems, owned)

	if len(orderItems) != 0 {
		t.Errorf("len(orderItems) = %d, want 0", len(orderItems))
	}
	if total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
}

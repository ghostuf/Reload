package handlers

import (
	"errors"
	"net/http"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
)

type CheckoutHandler struct {
	service     *services.CheckoutService
	paymentMode string
}

func NewCheckoutHandler(service *services.CheckoutService, paymentMode string) *CheckoutHandler {
	return &CheckoutHandler{service: service, paymentMode: paymentMode}
}

// Checkout is the mock, no-payment-processor flow. It stays registered even
// when PAYMENT_MODE=esewa -- a demo whose sandbox is down can still flip
// the env var back -- but is gated behind that flag so the two flows are
// never both live at once by accident.
func (h *CheckoutHandler) Checkout(c *gin.Context) {
	if h.paymentMode != "mock" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mock checkout is disabled; use /checkout/esewa/initiate"})
		return
	}

	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	order, err := h.service.Checkout(userID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCartEmpty), errors.Is(err, services.ErrNothingToPurchase):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "checkout failed"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

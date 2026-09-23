package handlers

import (
	"errors"
	"net/http"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
)

type CheckoutHandler struct {
	service *services.CheckoutService
}

func NewCheckoutHandler(service *services.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{service: service}
}

func (h *CheckoutHandler) Checkout(c *gin.Context) {
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

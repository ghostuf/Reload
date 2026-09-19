package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CartHandler struct {
	service *services.CartService
}

func NewCartHandler(service *services.CartService) *CartHandler {
	return &CartHandler{service: service}
}

type addCartItemRequest struct {
	GameID uint `json:"game_id"`
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	result, err := h.service.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result.Items,
		"total": result.Total,
	})
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req addCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.GameID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	item, err := h.service.AddToCart(userID, req.GameID)
	if err != nil {
		respondCartError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	gameID, err := strconv.ParseUint(c.Param("game_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	if err := h.service.RemoveFromCart(userID, uint(gameID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove cart item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"removed": true}})
}

// respondCartError maps service-level sentinel errors onto status codes, and
// keeps unexpected failures from leaking internals to the client.
func respondCartError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrAlreadyInCart):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add to cart"})
	}
}

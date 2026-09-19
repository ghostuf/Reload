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

type WishlistHandler struct {
	service *services.WishlistService
}

func NewWishlistHandler(service *services.WishlistService) *WishlistHandler {
	return &WishlistHandler{service: service}
}

type addWishlistItemRequest struct {
	GameID uint `json:"game_id"`
}

func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	items, err := h.service.GetWishlist(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *WishlistHandler) AddToWishlist(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req addWishlistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.GameID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	item, err := h.service.AddToWishlist(userID, req.GameID)
	if err != nil {
		respondWishlistError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *WishlistHandler) RemoveFromWishlist(c *gin.Context) {
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

	if err := h.service.RemoveFromWishlist(userID, uint(gameID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "wishlist item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove wishlist item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"removed": true}})
}

// respondWishlistError maps service-level sentinel errors onto status codes,
// and keeps unexpected failures from leaking internals to the client.
func respondWishlistError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrAlreadyInWishlist):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add to wishlist"})
	}
}

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GameHandler struct {
	service *services.GameService
}

func NewGameHandler(service *services.GameService) *GameHandler {
	return &GameHandler{service: service}
}

func (h *GameHandler) ListGames(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	result, err := h.service.ListGames(services.ListGamesParams{
		Search:   c.Query("search"),
		Genre:    c.Query("genre"),
		Sort:     c.Query("sort"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list games"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result.Games,
		"pagination": gin.H{
			"page":        result.Page,
			"page_size":   result.PageSize,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
	})
}

func (h *GameHandler) GetGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	game, err := h.service.GetGame(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": game})
}

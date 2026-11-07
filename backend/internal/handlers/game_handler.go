package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

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
		OnSale:   c.Query("on_sale") == "true",
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

// adminGameRequest is the admin-facing shape of a game: genres and gallery
// images as plain names and URLs, since an admin creating or editing a game
// has no way to already know genre IDs or image IDs.
type adminGameRequest struct {
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	Description    string   `json:"description"`
	Developer      string   `json:"developer"`
	Publisher      string   `json:"publisher"`
	ReleaseDate    string   `json:"release_date"`
	Price          float64  `json:"price"`
	DiscountPrice  *float64 `json:"discount_price"`
	DiscountEndsAt *string  `json:"discount_ends_at"`
	CoverImageURL  string   `json:"cover_image_url"`
	Genres         []string `json:"genres"`
	Images         []string `json:"images"`
}

// toGameInput parses the request's date strings and carries everything else
// through unchanged. A blank release_date is left as the zero time rather
// than rejected: it is not one of the fields the service treats as required.
func (r adminGameRequest) toGameInput() (services.GameInput, error) {
	input := services.GameInput{
		Title:         r.Title,
		Slug:          r.Slug,
		Description:   r.Description,
		Developer:     r.Developer,
		Publisher:     r.Publisher,
		Price:         r.Price,
		DiscountPrice: r.DiscountPrice,
		CoverImageURL: r.CoverImageURL,
		GenreNames:    r.Genres,
		ImageURLs:     r.Images,
	}

	if r.ReleaseDate != "" {
		releaseDate, err := time.Parse("2006-01-02", r.ReleaseDate)
		if err != nil {
			return services.GameInput{}, errInvalidDate
		}
		input.ReleaseDate = releaseDate
	}

	if r.DiscountEndsAt != nil && *r.DiscountEndsAt != "" {
		endsAt, err := time.Parse(time.RFC3339, *r.DiscountEndsAt)
		if err != nil {
			return services.GameInput{}, errInvalidDate
		}
		input.DiscountEndsAt = &endsAt
	}

	return input, nil
}

var errInvalidDate = errors.New("release_date must be YYYY-MM-DD and discount_ends_at must be RFC 3339")

func (h *GameHandler) CreateGame(c *gin.Context) {
	var req adminGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	input, err := req.toGameInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.service.CreateGame(input)
	if err != nil {
		respondGameError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": game})
}

func (h *GameHandler) UpdateGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	var req adminGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	input, err := req.toGameInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.service.UpdateGame(uint(id), input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}
		respondGameError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": game})
}

func (h *GameHandler) DeleteGame(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	if err := h.service.DeleteGame(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}

// respondGameError maps service-level sentinel errors onto status codes, and
// keeps unexpected failures from leaking internals to the client.
func respondGameError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrGameValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrSlugTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "game request failed"})
	}
}

package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReviewHandler struct {
	service *services.ReviewService
}

func NewReviewHandler(service *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

type reviewRequest struct {
	Rating int    `json:"rating"`
	Body   string `json:"body"`
}

// reviewResponse is what a review looks like to a reader: the reviewer's
// display name rather than their full user record, which would otherwise
// leak their email alongside every review they have left.
type reviewResponse struct {
	ID          uint      `json:"id"`
	GameID      uint      `json:"game_id"`
	Rating      int       `json:"rating"`
	Body        string    `json:"body"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *ReviewHandler) ListReviews(c *gin.Context) {
	gameID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.service.ListReviews(services.ListReviewsParams{
		GameID:   uint(gameID),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reviews"})
		return
	}

	responses := make([]reviewResponse, len(result.Reviews))
	for i, review := range result.Reviews {
		responses[i] = reviewResponse{
			ID:          review.ID,
			GameID:      review.GameID,
			Rating:      review.Rating,
			Body:        review.Body,
			DisplayName: review.User.DisplayName,
			CreatedAt:   review.CreatedAt,
			UpdatedAt:   review.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
		"pagination": gin.H{
			"page":        result.Page,
			"page_size":   result.PageSize,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
	})
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	gameID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	review, err := h.service.CreateReview(services.CreateReviewParams{
		UserID: userID,
		GameID: uint(gameID),
		Rating: req.Rating,
		Body:   req.Body,
	})
	if err != nil {
		respondReviewError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": gin.H{
		"id":         review.ID,
		"game_id":    review.GameID,
		"rating":     review.Rating,
		"body":       review.Body,
		"created_at": review.CreatedAt,
		"updated_at": review.UpdatedAt,
	}})
}

func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	review, err := h.service.UpdateReview(services.UpdateReviewParams{
		ReviewID: uint(reviewID),
		UserID:   userID,
		Rating:   req.Rating,
		Body:     req.Body,
	})
	if err != nil {
		respondReviewError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"id":         review.ID,
		"game_id":    review.GameID,
		"rating":     review.Rating,
		"body":       review.Body,
		"created_at": review.CreatedAt,
		"updated_at": review.UpdatedAt,
	}})
}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	if err := h.service.DeleteReview(uint(reviewID), userID); err != nil {
		respondReviewError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}

// respondReviewError maps service-level sentinel errors onto status codes,
// and keeps unexpected failures from leaking internals to the client.
func respondReviewError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrReviewValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrGameNotOwned):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrAlreadyReviewed):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrNotReviewAuthor):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "review request failed"})
	}
}

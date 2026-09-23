package handlers

import (
	"net/http"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	service *services.LibraryService
}

func NewLibraryHandler(service *services.LibraryService) *LibraryHandler {
	return &LibraryHandler{service: service}
}

func (h *LibraryHandler) GetLibrary(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	items, err := h.service.GetLibrary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load library"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

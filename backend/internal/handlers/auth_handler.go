package handlers

import (
	"errors"
	"net/http"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	service *services.AuthService
	cookies auth.CookieOptions
}

func NewAuthHandler(service *services.AuthService, cookies auth.CookieOptions) *AuthHandler {
	return &AuthHandler{service: service, cookies: cookies}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.service.Register(services.RegisterParams{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		respondAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, tokens, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		respondAuthError(c, err)
		return
	}

	auth.SetAuthCookies(c, h.cookies, tokens.AccessToken, tokens.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawToken, _ := auth.ReadRefreshCookie(c)

	tokens, err := h.service.Refresh(rawToken)
	if err != nil {
		// A rejected refresh means the session is over; drop the stale cookies
		// so the client is not stuck retrying with them.
		if errors.Is(err, services.ErrInvalidRefreshToken) {
			auth.ClearAuthCookies(c, h.cookies)
		}
		respondAuthError(c, err)
		return
	}

	auth.SetAuthCookies(c, h.cookies, tokens.AccessToken, tokens.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"refreshed": true}})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	rawToken, _ := auth.ReadRefreshCookie(c)

	if err := h.service.Logout(rawToken); err != nil {
		respondAuthError(c, err)
		return
	}

	auth.ClearAuthCookies(c, h.cookies)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"logged_out": true}})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := auth.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := h.service.GetUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Valid token for a user that no longer exists.
			auth.ClearAuthCookies(c, h.cookies)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// respondAuthError maps service-level sentinel errors onto status codes, and
// keeps unexpected failures from leaking internals to the client.
func respondAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidEmail),
		errors.Is(err, services.ErrWeakPassword),
		errors.Is(err, services.ErrDisplayNameTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrInvalidCredentials),
		errors.Is(err, services.ErrInvalidRefreshToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication request failed"})
	}
}

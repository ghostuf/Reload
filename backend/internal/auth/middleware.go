package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error codes returned alongside a 401 so the client can tell an expired
// session (worth a silent refresh) from a broken one (send the user to login).
const (
	CodeTokenExpired = "token_expired"
	CodeUnauthorized = "unauthorized"
)

// RequireAuth rejects any request without a valid access token, and stores the
// authenticated user's ID under ContextUserIDKey for the handlers behind it.
func RequireAuth(tm *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := authenticate(c, tm)
		if err != nil {
			code := CodeUnauthorized
			if errors.Is(err, ErrExpiredToken) {
				code = CodeTokenExpired
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not authenticated",
				"code":  code,
			})
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}

// OptionalAuth identifies the caller when it can, but lets anonymous requests
// through. Handlers use UserIDFromContext to tell the two apart — useful for
// routes that stay public but personalise their response when signed in.
func OptionalAuth(tm *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID, err := authenticate(c, tm); err == nil {
			c.Set(ContextUserIDKey, userID)
		}
		c.Next()
	}
}

// authenticate reads the access cookie and returns the user ID it carries.
func authenticate(c *gin.Context, tm *TokenManager) (uint, error) {
	token, ok := ReadAccessCookie(c)
	if !ok {
		return 0, ErrInvalidToken
	}
	return tm.ParseAccessToken(token)
}

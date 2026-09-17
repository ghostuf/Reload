package auth

import "github.com/gin-gonic/gin"

// ContextUserIDKey is the gin context key under which the authenticated user's
// ID is stored by the auth middleware.
const ContextUserIDKey = "user_id"

// UserIDFromContext returns the authenticated user's ID. The second result is
// false on any request that did not pass through the auth middleware.
func UserIDFromContext(c *gin.Context) (uint, bool) {
	value, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint)
	if !ok || userID == 0 {
		return 0, false
	}

	return userID, true
}

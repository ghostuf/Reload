package auth

import (
	"net/http"

	"github.com/ghostuf/reload/internal/models"
	"github.com/gin-gonic/gin"
)

// RoleStore looks up a user's role. AdminOnly depends on this rather than a
// concrete repository so it can be satisfied by anything that can answer the
// one question it needs; *repository.UserRepository is the production
// implementation.
type RoleStore interface {
	GetRole(userID uint) (models.UserRole, error)
}

// AdminOnly rejects any request whose authenticated user is not an admin. It
// must run after RequireAuth, which is what puts a user ID in the context for
// it to look up — the role lives in the database rather than the access
// token, so a demotion takes effect on the next request rather than waiting
// out the token's lifetime.
func AdminOnly(roles RoleStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := UserIDFromContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "not authenticated",
				"code":  CodeUnauthorized,
			})
			return
		}

		role, err := roles.GetRole(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin access"})
			return
		}

		if role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		c.Next()
	}
}

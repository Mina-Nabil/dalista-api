package middlewares

import (
	"net/http"

	userModels "dalista/dalista-be/app/repository/models/userModels"

	"github.com/gin-gonic/gin"
)

// CurrentUser returns the authenticated user set by AuthMiddleware, or (nil,
// false) if none is present.
func CurrentUser(c *gin.Context) (*userModels.User, bool) {
	raw, ok := c.Get(ContextUserKey)
	if !ok {
		return nil, false
	}
	user, ok := raw.(*userModels.User)
	return user, ok
}

// AbortIfCant checks the authenticated user holds the given permission,
// responding 403 and aborting if not. Must run after AuthMiddleware. Returns
// true when the request may proceed.
func AbortIfCant(c *gin.Context, permission string) bool {
	user, ok := CurrentUser(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}
	if !user.Can(permission) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return false
	}
	return true
}

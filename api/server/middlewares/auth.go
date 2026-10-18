package middlewares

import (
	"net/http"
	"strings"

	authServices "dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/app/utils"

	"github.com/gin-gonic/gin"
)

// ContextUserKey is the gin.Context key under which the authenticated *User is stored.
const ContextUserKey = "user"

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

// AuthMiddleware validates the access token, loads the user + permissions onto
// the context, and rejects inactive accounts with 403.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			unauthorized(c)
			return
		}
		userID, err := utils.ParseAccessToken(token)
		if err != nil {
			unauthorized(c)
			return
		}
		user, err := authServices.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if user == nil {
			unauthorized(c)
			return
		}
		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": authServices.AccountDeactivatedMessage})
			return
		}
		perms, err := authServices.LoadUserPermissions(c.Request.Context(), user.ID, user.Role)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		user.Permissions = perms

		c.Set(ContextUserKey, user)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
}

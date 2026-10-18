// Package authHandlers holds the HTTP handlers for authentication and session
// management: email/password, phone-OTP, Google/Facebook social sign-in, token
// refresh, logout, and post-signup phone verification.
package authHandlers

import (
	"net/http"

	"dalista/dalista-be/api/resources"
	userModels "dalista/dalista-be/app/repository/models/userModels"
	authServices "dalista/dalista-be/app/services/authServices"

	"github.com/gin-gonic/gin"
)

// writeAuth sends the standard auth success payload.
func writeAuth(c *gin.Context, status int, user *userModels.User, tokens authServices.TokenPair) {
	c.JSON(status, resources.AuthFromModel(user, tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresAt))
}

func badRequest(c *gin.Context, errs []string) {
	c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
}

func internalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
}

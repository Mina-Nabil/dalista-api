// Package billingHandlers holds the client-facing billing handlers: reading the
// user's subscription + storage, verifying Apple/Google purchases, and the
// admin subscription-oversight endpoints. Provider webhooks live in the
// appleIAPHandlers and playBillingHandlers packages.
package billingHandlers

import (
	"net/http"

	"dalista/dalista-be/api/server/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	user, ok := middlewares.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return uuid.UUID{}, false
	}
	return user.ID, true
}

func internalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
}

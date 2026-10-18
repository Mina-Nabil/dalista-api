package authHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/forms"
	authServices "dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/app/utils"

	"github.com/gin-gonic/gin"
)

// Refresh godoc
//
//	@Summary      Refresh tokens
//	@Description  Exchanges a valid refresh token for a new access + refresh token pair.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.RefreshForm  true  "Refresh token"
//	@Success      200      {object}  resources.AuthResource
//	@Failure      401      {object}  responses.ErrorResponse  "Invalid or expired refresh token"
//	@Router       /auth/refresh [post]
func Refresh(c *gin.Context) {
	var form forms.RefreshForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}

	user, tokens, invalidOrExpired, err := authServices.RefreshTokens(c.Request.Context(), form.RefreshToken)
	if err != nil {
		if errors.Is(err, authServices.ErrAccountDeactivated) {
			c.JSON(http.StatusForbidden, gin.H{"error": authServices.AccountDeactivatedMessage})
			return
		}
		internalError(c)
		return
	}
	if invalidOrExpired {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}
	writeAuth(c, http.StatusOK, user, tokens)
}

// Logout godoc
//
//	@Summary      Logout
//	@Description  Stateless logout. Tokens are not server-side revocable in v1; the client discards them. Endpoint exists for a consistent contract and future revocation.
//	@Tags         auth
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {object}  responses.ErrorResponse  "message: logged out"
//	@Router       /auth/logout [post]
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

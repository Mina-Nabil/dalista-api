package authHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/forms"
	authServices "dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/app/utils"
	"dalista/dalista-be/integrations/oauth"

	"github.com/gin-gonic/gin"
)

// GoogleAuth godoc
//
//	@Summary      Sign in / sign up with Google
//	@Description  Verifies a Google ID token server-side, then signs in or creates the account. Returns access + refresh tokens.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.GoogleAuthForm  true  "Google ID token"
//	@Success      200      {object}  resources.AuthResource
//	@Failure      401      {object}  responses.ErrorResponse  "Invalid token"
//	@Failure      503      {object}  responses.ErrorResponse  "Google OAuth not configured"
//	@Router       /auth/google [post]
func GoogleAuth(c *gin.Context) {
	var form forms.GoogleAuthForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}
	user, tokens, err := authServices.SignInWithGoogle(c.Request.Context(), form.IDToken)
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	writeAuth(c, http.StatusOK, user, tokens)
}

// FacebookAuth godoc
//
//	@Summary      Sign in / sign up with Facebook
//	@Description  Verifies a Facebook access token server-side, then signs in or creates the account. Returns access + refresh tokens.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.FacebookAuthForm  true  "Facebook access token"
//	@Success      200      {object}  resources.AuthResource
//	@Failure      401      {object}  responses.ErrorResponse  "Invalid token"
//	@Failure      503      {object}  responses.ErrorResponse  "Facebook OAuth not configured"
//	@Router       /auth/facebook [post]
func FacebookAuth(c *gin.Context) {
	var form forms.FacebookAuthForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}
	user, tokens, err := authServices.SignInWithFacebook(c.Request.Context(), form.AccessToken)
	if err != nil {
		writeOAuthError(c, err)
		return
	}
	writeAuth(c, http.StatusOK, user, tokens)
}

func writeOAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, oauth.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid social token"})
	case errors.Is(err, oauth.ErrNotConfigured):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Social sign-in is not configured"})
	case errors.Is(err, authServices.ErrAccountDeactivated):
		c.JSON(http.StatusForbidden, gin.H{"error": authServices.AccountDeactivatedMessage})
	default:
		internalError(c)
	}
}

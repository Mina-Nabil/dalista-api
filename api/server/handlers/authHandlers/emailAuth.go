package authHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/forms"
	authServices "dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/app/utils"

	"github.com/gin-gonic/gin"
)

// RegisterEmail godoc
//
//	@Summary      Sign up with email & password
//	@Description  Creates an email/password account and returns access + refresh tokens. Phone verification is a separate post-signup step.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.RegisterEmailForm  true  "Credentials"
//	@Success      201      {object}  resources.AuthResource
//	@Failure      400      {object}  responses.ValidationErrorResponse
//	@Failure      409      {object}  responses.ErrorResponse  "Email already registered"
//	@Router       /auth/register [post]
func RegisterEmail(c *gin.Context) {
	var form forms.RegisterEmailForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}
	if !utils.PasswordMeetsPolicy(form.Password) {
		badRequest(c, []string{"password must be at least 8 characters and include a letter and a digit"})
		return
	}

	user, tokens, err := authServices.RegisterEmail(c.Request.Context(), form.Email, form.Password)
	if err != nil {
		if errors.Is(err, authServices.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		internalError(c)
		return
	}
	writeAuth(c, http.StatusCreated, user, tokens)
}

// LoginEmail godoc
//
//	@Summary      Sign in with email & password
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.LoginEmailForm  true  "Credentials"
//	@Success      200      {object}  resources.AuthResource
//	@Failure      400      {object}  responses.ValidationErrorResponse
//	@Failure      401      {object}  responses.ErrorResponse  "Invalid credentials"
//	@Failure      403      {object}  responses.ErrorResponse  "Account deactivated"
//	@Router       /auth/login [post]
func LoginEmail(c *gin.Context) {
	var form forms.LoginEmailForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}

	user, tokens, err := authServices.LoginEmail(c.Request.Context(), form.Email, form.Password)
	if err != nil {
		switch {
		case errors.Is(err, authServices.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		case errors.Is(err, authServices.ErrAccountDeactivated):
			c.JSON(http.StatusForbidden, gin.H{"error": authServices.AccountDeactivatedMessage})
		default:
			internalError(c)
		}
		return
	}
	writeAuth(c, http.StatusOK, user, tokens)
}

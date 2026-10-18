package authHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	authServices "dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/app/utils"

	"github.com/gin-gonic/gin"
)

// RequestSigninOTP godoc
//
//	@Summary      Request a phone sign-in OTP
//	@Description  Sends a one-time code to the phone number. No account required — a new one is created on verify.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.PhoneForm  true  "Phone number (E.164)"
//	@Success      200      {object}  responses.ErrorResponse  "message: code sent"
//	@Failure      400      {object}  responses.ValidationErrorResponse
//	@Router       /auth/otp/request [post]
func RequestSigninOTP(c *gin.Context) {
	var form forms.PhoneForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}
	if err := authServices.RequestSigninOTP(c.Request.Context(), form.Phone); err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "code sent"})
}

// VerifySigninOTP godoc
//
//	@Summary      Verify a phone sign-in OTP
//	@Description  Verifies the code and signs in (creating the account on first use). Returns access + refresh tokens.
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.VerifyOTPForm  true  "Phone + code"
//	@Success      200      {object}  resources.AuthResource
//	@Failure      400      {object}  responses.ValidationErrorResponse
//	@Failure      401      {object}  responses.ErrorResponse  "Invalid code"
//	@Failure      429      {object}  responses.ErrorResponse  "Too many attempts"
//	@Router       /auth/otp/verify [post]
func VerifySigninOTP(c *gin.Context) {
	var form forms.VerifyOTPForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}
	user, tokens, err := authServices.VerifySigninOTP(c.Request.Context(), form.Phone, form.Code)
	if err != nil {
		writeOTPError(c, err)
		return
	}
	writeAuth(c, http.StatusOK, user, tokens)
}

// RequestPhoneVerification godoc
//
//	@Summary      Request a phone-verification OTP (post-signup)
//	@Description  For an authenticated user who signed up via email/Google/Facebook — sends an OTP to the phone they want to verify.
//	@Tags         auth
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.PhoneForm  true  "Phone number (E.164)"
//	@Success      200      {object}  responses.ErrorResponse  "message: code sent"
//	@Failure      409      {object}  responses.ErrorResponse  "Phone already in use"
//	@Router       /auth/phone/request [post]
func RequestPhoneVerification(c *gin.Context) {
	user, ok := middlewares.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	var form forms.PhoneForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}
	if err := authServices.RequestVerifyOTP(c.Request.Context(), user.ID, form.Phone); err != nil {
		if errors.Is(err, authServices.ErrPhoneTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Phone already in use"})
			return
		}
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "code sent"})
}

// VerifyPhone godoc
//
//	@Summary      Verify a phone number (post-signup)
//	@Description  Confirms the OTP and marks the authenticated user's phone verified.
//	@Tags         auth
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.VerifyOTPForm  true  "Phone + code"
//	@Success      200      {object}  resources.ProfileResource
//	@Failure      401      {object}  responses.ErrorResponse  "Invalid code"
//	@Failure      409      {object}  responses.ErrorResponse  "Phone already in use"
//	@Router       /auth/phone/verify [post]
func VerifyPhone(c *gin.Context) {
	user, ok := middlewares.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	var form forms.VerifyOTPForm
	if err := c.ShouldBindJSON(&form); err != nil {
		badRequest(c, utils.FormatValidationErrors(err))
		return
	}
	updated, err := authServices.VerifyPhone(c.Request.Context(), user.ID, form.Phone, form.Code)
	if err != nil {
		if errors.Is(err, authServices.ErrPhoneTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "Phone already in use"})
			return
		}
		writeOTPError(c, err)
		return
	}
	_, profile, err := authServices.GetProfile(c.Request.Context(), updated.ID)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, resources.ProfileFromModel(updated, profile))
}

// writeOTPError maps OTP service errors to HTTP responses.
func writeOTPError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, authServices.ErrOTPNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active verification code; request a new one"})
	case errors.Is(err, authServices.ErrOTPInvalid):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid verification code"})
	case errors.Is(err, authServices.ErrOTPTooManyAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many attempts; request a new code"})
	case errors.Is(err, authServices.ErrAccountDeactivated):
		c.JSON(http.StatusForbidden, gin.H{"error": authServices.AccountDeactivatedMessage})
	default:
		internalError(c)
	}
}

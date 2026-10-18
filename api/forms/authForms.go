package forms

// RegisterEmailForm is the body for email/password sign-up.
type RegisterEmailForm struct {
	Email    string `json:"email" binding:"required,email" example:"jane@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"hunter2go"`
}

// LoginEmailForm is the body for email/password sign-in.
type LoginEmailForm struct {
	Email    string `json:"email" binding:"required,email" example:"jane@example.com"`
	Password string `json:"password" binding:"required" example:"hunter2go"`
}

// RefreshForm is the body for the token-refresh endpoint.
type RefreshForm struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// PhoneForm requests an OTP for a phone number (sign-in or verification).
type PhoneForm struct {
	Phone string `json:"phone" binding:"required,e164" example:"+201001234567"`
}

// VerifyOTPForm submits a phone + OTP code.
type VerifyOTPForm struct {
	Phone string `json:"phone" binding:"required,e164" example:"+201001234567"`
	Code  string `json:"code" binding:"required,len=6" example:"004213"`
}

// GoogleAuthForm carries a Google ID token for server-side verification.
type GoogleAuthForm struct {
	IDToken string `json:"id_token" binding:"required"`
}

// FacebookAuthForm carries a Facebook user access token for server-side verification.
type FacebookAuthForm struct {
	AccessToken string `json:"access_token" binding:"required"`
}

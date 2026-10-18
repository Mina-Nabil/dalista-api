package resources

import (
	"time"

	userModels "dalista/dalista-be/app/repository/models/userModels"
)

// UserResource is the public representation of a user.
type UserResource struct {
	ID            string  `json:"id"`
	Email         *string `json:"email"`
	Phone         *string `json:"phone"`
	Role          string  `json:"role"`
	PhoneVerified bool    `json:"phone_verified"`
	IsActive      bool    `json:"is_active"`
}

// UserFromModel maps a User model to its resource.
func UserFromModel(u *userModels.User) UserResource {
	return UserResource{
		ID:            u.ID.String(),
		Email:         u.Email,
		Phone:         u.Phone,
		Role:          string(u.Role),
		PhoneVerified: u.PhoneVerified,
		IsActive:      u.IsActive,
	}
}

// AuthResource is the success payload for every sign-in/sign-up/refresh response.
type AuthResource struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type" example:"Bearer"`
	ExpiresIn    int64        `json:"expires_in" example:"28800"`
	User         UserResource `json:"user"`
}

// AuthFromModel builds the auth payload from a user and an issued token pair.
func AuthFromModel(u *userModels.User, accessToken, refreshToken string, expiresAt time.Time) AuthResource {
	return AuthResource{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(expiresAt).Seconds()),
		User:         UserFromModel(u),
	}
}

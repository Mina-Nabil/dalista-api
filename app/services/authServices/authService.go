package authServices

import (
	"context"
	"errors"
	"strings"
	"time"

	"dalista/dalista-be/app/utils"

	userModels "dalista/dalista-be/app/repository/models/userModels"
)

// AccountDeactivatedMessage is the API message for inactive accounts.
const AccountDeactivatedMessage = "Account has been deactivated"

var (
	// ErrEmailTaken is returned when registering an email that already exists.
	ErrEmailTaken = errors.New("email already registered")
	// ErrInvalidCredentials is returned for a bad email/password pair.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrAccountDeactivated is returned when a valid login targets an inactive account.
	ErrAccountDeactivated = errors.New(AccountDeactivatedMessage)
)

// TokenPair is an access + refresh token result.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// IssueTokens mints an access + refresh token pair for the user.
func IssueTokens(user *userModels.User) (TokenPair, error) {
	access, expiry, err := utils.GenerateAccessToken(user)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, _, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresAt: expiry}, nil
}

// RegisterEmail creates a new email/password account and returns it with a
// fresh token pair. Phone verification is a separate post-signup step.
func RegisterEmail(ctx context.Context, email, password string) (*userModels.User, TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existing, err := GetUserByEmail(ctx, email)
	if err != nil {
		return nil, TokenPair{}, err
	}
	if existing != nil {
		return nil, TokenPair{}, ErrEmailTaken
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, TokenPair{}, err
	}

	displayName := displayNameFromEmail(email)
	user, err := createUserWithIdentity(ctx, &email, nil, false,
		userModels.ProviderEmail, email, &hash, &displayName, nil)
	if err != nil {
		return nil, TokenPair{}, err
	}

	tokens, err := IssueTokens(user)
	return user, tokens, err
}

// LoginEmail authenticates an email/password pair.
func LoginEmail(ctx context.Context, email, password string) (*userModels.User, TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	identity, err := GetIdentity(ctx, userModels.ProviderEmail, email)
	if err != nil {
		return nil, TokenPair{}, err
	}
	if identity == nil || identity.PasswordHash == nil {
		return nil, TokenPair{}, ErrInvalidCredentials
	}
	if !utils.VerifyPassword(password, *identity.PasswordHash) {
		return nil, TokenPair{}, ErrInvalidCredentials
	}

	user, err := GetUserByID(ctx, identity.UserID.String())
	if err != nil {
		return nil, TokenPair{}, err
	}
	if user == nil {
		return nil, TokenPair{}, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, TokenPair{}, ErrAccountDeactivated
	}

	tokens, err := IssueTokens(user)
	return user, tokens, err
}

// RefreshTokens validates a refresh token and issues a new token pair.
// invalidOrExpired is true when the token can't be trusted (→ 401).
func RefreshTokens(ctx context.Context, refreshToken string) (user *userModels.User, tokens TokenPair, invalidOrExpired bool, err error) {
	sub, perr := utils.ParseRefreshToken(refreshToken)
	if perr != nil {
		return nil, TokenPair{}, true, nil
	}
	user, err = GetUserByID(ctx, sub)
	if err != nil {
		return nil, TokenPair{}, false, err
	}
	if user == nil {
		return nil, TokenPair{}, true, nil
	}
	if !user.IsActive {
		return nil, TokenPair{}, false, ErrAccountDeactivated
	}
	tokens, err = IssueTokens(user)
	return user, tokens, false, err
}

func displayNameFromEmail(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
}

package authServices

import (
	"context"
	"strings"

	userModels "dalista/dalista-be/app/repository/models/userModels"
	"dalista/dalista-be/integrations/oauth"
)

// SignInWithGoogle verifies a Google ID token and signs in (or creates) the user.
func SignInWithGoogle(ctx context.Context, idToken string) (*userModels.User, TokenPair, error) {
	ext, err := oauth.VerifyGoogle(ctx, idToken)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return upsertSocialUser(ctx, userModels.ProviderGoogle, ext)
}

// SignInWithFacebook verifies a Facebook access token and signs in (or creates) the user.
func SignInWithFacebook(ctx context.Context, accessToken string) (*userModels.User, TokenPair, error) {
	ext, err := oauth.VerifyFacebook(ctx, accessToken)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return upsertSocialUser(ctx, userModels.ProviderFacebook, ext)
}

// upsertSocialUser resolves a verified external identity to a Dalista user:
//   - existing identity → sign in
//   - no identity but matching email → link the provider to that account
//   - otherwise → create a new account
func upsertSocialUser(ctx context.Context, provider userModels.Provider, ext *oauth.ExternalIdentity) (*userModels.User, TokenPair, error) {
	identity, err := GetIdentity(ctx, provider, ext.Subject)
	if err != nil {
		return nil, TokenPair{}, err
	}

	var user *userModels.User
	switch {
	case identity != nil:
		user, err = GetUserByID(ctx, identity.UserID.String())
		if err != nil {
			return nil, TokenPair{}, err
		}
		if user == nil {
			return nil, TokenPair{}, ErrInvalidCredentials
		}

	default:
		email := strings.ToLower(strings.TrimSpace(ext.Email))
		if email != "" {
			if existing, lerr := GetUserByEmail(ctx, email); lerr != nil {
				return nil, TokenPair{}, lerr
			} else if existing != nil {
				// Link this social provider to the existing email account.
				if err := linkIdentity(ctx, existing.ID, provider, ext.Subject, nil); err != nil {
					return nil, TokenPair{}, err
				}
				user = existing
			}
		}
		if user == nil {
			var emailPtr *string
			if email != "" {
				emailPtr = &email
			}
			var namePtr, picPtr *string
			if ext.Name != "" {
				namePtr = &ext.Name
			}
			if ext.PictureURL != "" {
				picPtr = &ext.PictureURL
			}
			user, err = createUserWithIdentity(ctx, emailPtr, nil, false, provider, ext.Subject, nil, namePtr, picPtr)
			if err != nil {
				return nil, TokenPair{}, err
			}
		}
	}

	if !user.IsActive {
		return nil, TokenPair{}, ErrAccountDeactivated
	}

	tokens, err := IssueTokens(user)
	return user, tokens, err
}

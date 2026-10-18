// Package oauth verifies social sign-in tokens server-side (Google ID tokens,
// Facebook access tokens) and returns a normalized external identity. The
// client never trusts its own login — it sends the provider token, which the
// backend verifies here before issuing a Dalista JWT.
package oauth

import "errors"

// ExternalIdentity is the normalized result of verifying a social token.
type ExternalIdentity struct {
	Subject    string // stable provider user id → AuthIdentity.ProviderSubject
	Email      string // may be empty if the user withheld email permission
	Name       string
	PictureURL string
}

// ErrInvalidToken means the provider rejected the token or it failed our
// audience/app checks. Surfaced to callers as a 401.
var ErrInvalidToken = errors.New("invalid social token")

// ErrNotConfigured means the relevant provider credentials are missing.
var ErrNotConfigured = errors.New("oauth provider not configured")

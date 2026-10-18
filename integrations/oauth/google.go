package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dalista/dalista-be/config"
)

const googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

// VerifyGoogle validates a Google ID token via Google's tokeninfo endpoint and
// checks the audience matches our OAuth client id. Returns the normalized
// identity (sub, email, name, picture).
func VerifyGoogle(ctx context.Context, idToken string) (*ExternalIdentity, error) {
	cfg := config.GetConfig()
	if strings.TrimSpace(cfg.GoogleOAuthClientID) == "" {
		return nil, ErrNotConfigured
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	u := googleTokenInfoURL + "?" + url.Values{"id_token": {idToken}}.Encode()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidToken
	}

	var info struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Exp           string `json:"exp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("google: decode tokeninfo: %w", err)
	}

	// Audience must match our client id, otherwise this token was minted for a
	// different app and must not be trusted.
	if info.Aud != cfg.GoogleOAuthClientID || info.Sub == "" {
		return nil, ErrInvalidToken
	}

	return &ExternalIdentity{
		Subject:    info.Sub,
		Email:      info.Email,
		Name:       info.Name,
		PictureURL: info.Picture,
	}, nil
}

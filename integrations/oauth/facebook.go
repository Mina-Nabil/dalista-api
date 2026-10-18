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

const (
	fbDebugTokenURL = "https://graph.facebook.com/debug_token"
	fbMeURL         = "https://graph.facebook.com/me"
)

// VerifyFacebook validates a Facebook user access token in two steps:
//  1. debug_token (authenticated with the app access token) confirms the token
//     was issued for *our* app and is valid.
//  2. /me fetches the profile (id, name, email, picture).
func VerifyFacebook(ctx context.Context, accessToken string) (*ExternalIdentity, error) {
	cfg := config.GetConfig()
	if strings.TrimSpace(cfg.FacebookAppID) == "" || strings.TrimSpace(cfg.FacebookAppSecret) == "" {
		return nil, ErrNotConfigured
	}
	appToken := cfg.FacebookAppID + "|" + cfg.FacebookAppSecret

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 1. Verify the token belongs to our app and is valid.
	debugURL := fbDebugTokenURL + "?" + url.Values{
		"input_token":  {accessToken},
		"access_token": {appToken},
	}.Encode()
	var debug struct {
		Data struct {
			AppID   string `json:"app_id"`
			IsValid bool   `json:"is_valid"`
			UserID  string `json:"user_id"`
		} `json:"data"`
	}
	if err := fbGet(reqCtx, debugURL, &debug); err != nil {
		return nil, err
	}
	if !debug.Data.IsValid || debug.Data.AppID != cfg.FacebookAppID || debug.Data.UserID == "" {
		return nil, ErrInvalidToken
	}

	// 2. Fetch the profile.
	meURL := fbMeURL + "?" + url.Values{
		"fields":       {"id,name,email,picture"},
		"access_token": {accessToken},
	}.Encode()
	var me struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	if err := fbGet(reqCtx, meURL, &me); err != nil {
		return nil, err
	}
	if me.ID == "" || me.ID != debug.Data.UserID {
		return nil, ErrInvalidToken
	}

	return &ExternalIdentity{
		Subject:    me.ID,
		Email:      me.Email,
		Name:       me.Name,
		PictureURL: me.Picture.Data.URL,
	}, nil
}

func fbGet(ctx context.Context, u string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ErrInvalidToken
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("facebook: decode response: %w", err)
	}
	return nil
}

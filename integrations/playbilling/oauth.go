package playbilling

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dalista/dalista-be/config"

	"github.com/golang-jwt/jwt/v5"
)

const androidPublisherScope = "https://www.googleapis.com/auth/androidpublisher"

type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// accessToken exchanges a signed service-account JWT for an OAuth2 access token
// with the androidpublisher scope (JWT bearer grant). Tokens are short-lived;
// callers fetch one per request (fine for our low webhook volume).
func accessToken(ctx context.Context) (string, error) {
	cfg := config.GetConfig()
	if strings.TrimSpace(cfg.GooglePlayCredentialsJSON) == "" {
		return "", ErrNotConfigured
	}
	var sa serviceAccount
	if err := json.Unmarshal([]byte(cfg.GooglePlayCredentialsJSON), &sa); err != nil {
		return "", fmt.Errorf("google: parse service account: %w", err)
	}
	tokenURI := sa.TokenURI
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}

	key, err := parseRSAKey(sa.PrivateKey)
	if err != nil {
		return "", err
	}

	now := time.Now()
	assertion := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   sa.ClientEmail,
		"scope": androidPublisherScope,
		"aud":   tokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(30 * time.Minute).Unix(),
	})
	signed, err := assertion.SignedString(key)
	if err != nil {
		return "", err
	}

	form := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {signed},
	}
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google: token exchange failed: status %d", resp.StatusCode)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("google: empty access token")
	}
	return out.AccessToken, nil
}

func parseRSAKey(pemStr string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("google: invalid private key PEM")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("google: parse private key: %w", err)
	}
	return key, nil
}

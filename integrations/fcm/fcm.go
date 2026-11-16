// Package fcm sends push notifications via the Firebase Cloud Messaging HTTP v1
// API, authenticated with the FCM service-account credentials (JWT bearer grant
// → OAuth2 access token, same hand-rolled approach as the Play Billing client —
// no heavy SDK). In local dev, with no credentials configured, a send is logged
// instead of delivered so the rest of the flow works offline.
package fcm

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/config"

	"github.com/golang-jwt/jwt/v5"
)

const (
	messagingScope = "https://www.googleapis.com/auth/firebase.messaging"
	sendURLFormat  = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
)

var (
	// ErrNotConfigured means FCM credentials are absent (non-dev environments).
	ErrNotConfigured = errors.New("fcm: credentials not configured")
	// ErrInvalidToken means FCM rejected the device token as unregistered/invalid,
	// so the caller should prune it.
	ErrInvalidToken = errors.New("fcm: device token invalid")
)

// Message is a single push to one device token.
type Message struct {
	Token string
	Title string
	Body  string
	Data  map[string]string
}

type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
	ProjectID   string `json:"project_id"`
}

// Send delivers one message. In dev (no credentials) it logs and returns nil.
// On an unregistered/invalid token it returns ErrInvalidToken.
func Send(ctx context.Context, m Message) error {
	cfg := config.GetConfig()
	raw := strings.TrimSpace(cfg.FCMCredentialsJSON)
	if raw == "" {
		if cfg.IsStagingEnvironment() {
			logger.Warning(ctx, "FCM not configured — logging push instead of sending",
				logger.F("token", truncate(m.Token)), logger.F("title", m.Title), logger.F("body", m.Body))
			return nil
		}
		return ErrNotConfigured
	}

	var sa serviceAccount
	if err := json.Unmarshal([]byte(raw), &sa); err != nil {
		return fmt.Errorf("fcm: parse service account: %w", err)
	}
	if sa.ProjectID == "" {
		return fmt.Errorf("fcm: service account missing project_id")
	}
	token, err := accessToken(ctx, &sa)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token":        m.Token,
			"notification": map[string]string{"title": m.Title, "body": m.Body},
			"data":         m.Data,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost,
		fmt.Sprintf(sendURLFormat, sa.ProjectID), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	// 404 (UNREGISTERED) or 400 (invalid token) → caller prunes the token.
	if resp.StatusCode == http.StatusNotFound ||
		(resp.StatusCode == http.StatusBadRequest && strings.Contains(string(respBody), "token")) {
		return ErrInvalidToken
	}
	return fmt.Errorf("fcm: send failed: status %d: %s", resp.StatusCode, truncate(string(respBody)))
}

// accessToken mints an OAuth2 access token for the firebase.messaging scope via
// a signed service-account JWT (JWT bearer grant).
func accessToken(ctx context.Context, sa *serviceAccount) (string, error) {
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
		"scope": messagingScope,
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
		return "", fmt.Errorf("fcm: token exchange failed: status %d", resp.StatusCode)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("fcm: empty access token")
	}
	return out.AccessToken, nil
}

func parseRSAKey(pemStr string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("fcm: invalid private key PEM")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("fcm: parse private key: %w", err)
	}
	return key, nil
}

func truncate(s string) string {
	if len(s) > 32 {
		return s[:32] + "…"
	}
	return s
}

package appstore

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"dalista/dalista-be/config"

	"github.com/golang-jwt/jwt/v5"
)

const (
	hostProduction = "https://api.storekit.itunes.apple.com"
	hostSandbox    = "https://api.storekit-sandbox.itunes.apple.com"
)

// apiHost returns the App Store Server API host for the configured environment.
func apiHost() string {
	if strings.EqualFold(config.GetConfig().AppleIAPEnvironment, "production") {
		return hostProduction
	}
	return hostSandbox
}

// generateAPIToken mints the ES256 bearer token for App Store Server API calls,
// signed with the App Store Connect API key (.p8). Valid for up to 1 hour.
func generateAPIToken() (string, error) {
	cfg := config.GetConfig()
	if cfg.AppleIAPKeyID == "" || cfg.AppleIAPIssuerID == "" || strings.TrimSpace(cfg.AppleIAPPrivateKey) == "" {
		return "", ErrNotConfigured
	}
	key, err := parseP8(cfg.AppleIAPPrivateKey)
	if err != nil {
		return "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": cfg.AppleIAPIssuerID,
		"iat": now.Unix(),
		"exp": now.Add(30 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": cfg.AppleIAPBundleID,
	})
	token.Header["kid"] = cfg.AppleIAPKeyID
	return token.SignedString(key)
}

func parseP8(pemStr string) (interface{}, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("apple: invalid .p8 PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apple: parse .p8: %w", err)
	}
	return key, nil
}

// GetTransactionInfo fetches and verifies the latest transaction for a
// transaction id via the App Store Server API, returning the decoded info.
func GetTransactionInfo(ctx context.Context, transactionID string) (*TransactionInfo, error) {
	token, err := generateAPIToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/inApps/v1/transactions/%s", apiHost(), transactionID)
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apple: transaction lookup failed: status %d", resp.StatusCode)
	}

	// The response body is {"signedTransactionInfo": "<JWS>"}.
	var body struct {
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("apple: decode transaction response: %w", err)
	}
	return DecodeTransaction(body.SignedTransactionInfo)
}

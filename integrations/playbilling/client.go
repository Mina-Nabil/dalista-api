package playbilling

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

const apiBase = "https://androidpublisher.googleapis.com/androidpublisher/v3"

// VerifySubscription verifies a purchase token against the Play Developer API
// (subscriptions.get) and returns the authoritative purchase state. This is the
// real trust boundary — the client is never believed about its own purchase.
func VerifySubscription(ctx context.Context, subscriptionID, purchaseToken string) (*SubscriptionPurchase, error) {
	cfg := config.GetConfig()
	if strings.TrimSpace(cfg.GooglePlayPackageName) == "" {
		return nil, ErrNotConfigured
	}
	token, err := accessToken(ctx)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/applications/%s/purchases/subscriptions/%s/tokens/%s",
		apiBase,
		url.PathEscape(cfg.GooglePlayPackageName),
		url.PathEscape(subscriptionID),
		url.PathEscape(purchaseToken),
	)

	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
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
		return nil, fmt.Errorf("google: subscription verify failed: status %d", resp.StatusCode)
	}

	var purchase SubscriptionPurchase
	if err := json.NewDecoder(resp.Body).Decode(&purchase); err != nil {
		return nil, fmt.Errorf("google: decode subscription: %w", err)
	}
	return &purchase, nil
}

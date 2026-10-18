// Package appstore is the Apple App Store Server API client: App Store Server
// API transaction lookups (receipt validation) and App Store Server
// Notifications V2 (JWS) decoding + signature verification. See the technical
// architecture doc §2.6. No Stripe-style shared-secret webhook here — Apple
// signs payloads with JWS, verified against the x5c certificate chain.
package appstore

import "errors"

var (
	// ErrNotConfigured means the App Store Connect API credentials are missing.
	ErrNotConfigured = errors.New("apple app store not configured")
	// ErrInvalidSignature means a JWS payload failed signature/chain verification.
	ErrInvalidSignature = errors.New("invalid apple jws signature")
)

// TransactionInfo is the decoded JWSTransaction payload (subset we use).
type TransactionInfo struct {
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	ProductID             string `json:"productId"`
	// ExpiresDate is epoch milliseconds (0 for non-renewing).
	ExpiresDate int64  `json:"expiresDate"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
	BundleID    string `json:"bundleId"`
	// RevocationDate is set (epoch ms) when a purchase was refunded/revoked.
	RevocationDate int64 `json:"revocationDate"`
}

// NotificationData is the `data` object of a V2 notification.
type NotificationData struct {
	BundleID              string `json:"bundleId"`
	Environment           string `json:"environment"`
	SignedTransactionInfo string `json:"signedTransactionInfo"`
	SignedRenewalInfo     string `json:"signedRenewalInfo"`
}

// Notification is the decoded responseBodyV2DecodedPayload (subset).
// NotificationType / Subtype drive the subscription-state sync. See:
// https://developer.apple.com/documentation/appstoreservernotifications
type Notification struct {
	NotificationType string           `json:"notificationType"`
	Subtype          string           `json:"subtype"`
	NotificationUUID string           `json:"notificationUUID"`
	Data             NotificationData `json:"data"`
}

package appstore

import "time"

// ParseNotification verifies and decodes an App Store Server Notification V2
// `signedPayload` (the top-level JWS delivered to the webhook).
func ParseNotification(signedPayload string) (*Notification, error) {
	var n Notification
	if err := verifyAndDecodeJWS(signedPayload, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

// DecodeTransaction verifies and decodes a signed transaction JWS (either the
// `signedTransactionInfo` inside a notification or the App Store Server API's
// transaction response).
func DecodeTransaction(signedTransactionInfo string) (*TransactionInfo, error) {
	var t TransactionInfo
	if err := verifyAndDecodeJWS(signedTransactionInfo, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// ExpiresAt converts the millisecond ExpiresDate into a *time.Time (nil for 0).
func (t *TransactionInfo) ExpiresAt() *time.Time {
	if t.ExpiresDate == 0 {
		return nil
	}
	tm := time.UnixMilli(t.ExpiresDate)
	return &tm
}

// Revoked reports whether the transaction was refunded/revoked.
func (t *TransactionInfo) Revoked() bool { return t.RevocationDate != 0 }

// Package playbilling is the Google Play Developer API client: server-side
// purchase/subscription verification and Real-time Developer Notifications
// (RTDN, delivered via Pub/Sub push) decoding. See technical architecture §2.6.
package playbilling

import (
	"errors"
	"strconv"
	"time"
)

// ErrNotConfigured means the Play service-account credentials are missing.
var ErrNotConfigured = errors.New("google play not configured")

// SubscriptionPurchase is the Play Developer API v1 subscriptions.get response
// (subset). paymentState: 0 pending, 1 received, 2 free trial, 3 deferred.
type SubscriptionPurchase struct {
	ExpiryTimeMillis     string `json:"expiryTimeMillis"`
	StartTimeMillis      string `json:"startTimeMillis"`
	AutoRenewing         bool   `json:"autoRenewing"`
	OrderID              string `json:"orderId"`
	PaymentState         *int   `json:"paymentState"`
	AcknowledgementState int    `json:"acknowledgementState"`
	// CancelReason is set when the subscription was canceled.
	CancelReason *int `json:"cancelReason"`
}

// ExpiryTime parses ExpiryTimeMillis into a *time.Time (nil when absent/invalid).
func (s *SubscriptionPurchase) ExpiryTime() *time.Time {
	if s.ExpiryTimeMillis == "" {
		return nil
	}
	ms, err := strconv.ParseInt(s.ExpiryTimeMillis, 10, 64)
	if err != nil {
		return nil
	}
	t := time.UnixMilli(ms)
	return &t
}

// RTDN notification type codes (subscriptionNotification.notificationType).
const (
	RTDNRecovered            = 1
	RTDNRenewed              = 2
	RTDNCanceled             = 3
	RTDNPurchased            = 4
	RTDNOnHold               = 5
	RTDNInGracePeriod        = 6
	RTDNRestarted            = 7
	RTDNPriceChangeConfirmed = 8
	RTDNDeferred             = 9
	RTDNPaused               = 10
	RTDNRevoked              = 12
	RTDNExpired              = 13
)

// SubscriptionNotification is the subscription portion of an RTDN.
type SubscriptionNotification struct {
	Version          string `json:"version"`
	NotificationType int    `json:"notificationType"`
	PurchaseToken    string `json:"purchaseToken"`
	SubscriptionID   string `json:"subscriptionId"`
}

// DeveloperNotification is the decoded RTDN payload.
type DeveloperNotification struct {
	Version                  string                    `json:"version"`
	PackageName              string                    `json:"packageName"`
	EventTimeMillis          string                    `json:"eventTimeMillis"`
	SubscriptionNotification *SubscriptionNotification `json:"subscriptionNotification"`
	TestNotification         *struct {
		Version string `json:"version"`
	} `json:"testNotification"`
}

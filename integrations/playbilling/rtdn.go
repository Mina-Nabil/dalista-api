package playbilling

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// PubSubPushBody is the envelope Google Pub/Sub delivers to a push endpoint.
type PubSubPushBody struct {
	Message struct {
		Data      string `json:"data"` // base64-encoded DeveloperNotification JSON
		MessageID string `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// DecodeRTDN unpacks a Pub/Sub push envelope into the DeveloperNotification.
//
// Note on authenticity: the authoritative check is re-verifying the purchase
// token against the Play Developer API (VerifySubscription) on every event —
// the notification itself only tells us *which* token changed. Pub/Sub push
// OIDC-token verification (audience/issuer) is a recommended transport-level
// hardening and is tracked for the security review (DAL-164).
func DecodeRTDN(raw []byte) (*DeveloperNotification, error) {
	var env PubSubPushBody
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("google rtdn: decode envelope: %w", err)
	}
	if env.Message.Data == "" {
		return nil, fmt.Errorf("google rtdn: empty message data")
	}
	decoded, err := base64.StdEncoding.DecodeString(env.Message.Data)
	if err != nil {
		return nil, fmt.Errorf("google rtdn: decode data: %w", err)
	}
	var notif DeveloperNotification
	if err := json.Unmarshal(decoded, &notif); err != nil {
		return nil, fmt.Errorf("google rtdn: decode notification: %w", err)
	}
	return &notif, nil
}

// StatusForNotificationType maps an RTDN type to whether it still grants
// entitlement, so the caller can set the subscription status. Returns
// (entitled, recognized).
func StatusForNotificationType(t int) (entitled bool, recognized bool) {
	switch t {
	case RTDNPurchased, RTDNRenewed, RTDNRecovered, RTDNRestarted, RTDNInGracePeriod, RTDNPriceChangeConfirmed, RTDNDeferred:
		return true, true
	case RTDNCanceled, RTDNOnHold, RTDNPaused, RTDNRevoked, RTDNExpired:
		return false, true
	default:
		return false, false
	}
}

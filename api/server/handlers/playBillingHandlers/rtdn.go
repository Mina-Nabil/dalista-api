// Package playBillingHandlers handles Google Play Real-time Developer
// Notifications, delivered as Pub/Sub push messages, for renewals,
// cancellations, and refunds. Authenticity rests on re-verifying the purchase
// token against the Play Developer API on every event.
package playBillingHandlers

import (
	"io"
	"net/http"
	"time"

	"dalista/dalista-be/api/server/logger"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"
	billingServices "dalista/dalista-be/app/services/billingServices"
	"dalista/dalista-be/integrations/playbilling"

	"github.com/gin-gonic/gin"
)

// RTDN godoc
//
//	@Summary	Google Play Real-time Developer Notifications (Pub/Sub push)
//	@Description	Decodes the Pub/Sub envelope, re-verifies the purchase token against the Play Developer API (the real trust boundary), then syncs subscription state. Public endpoint.
//	@Tags		billing-webhooks
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	responses.ErrorResponse	"acknowledged"
//	@Failure	400	{object}	responses.ErrorResponse
//	@Router		/billing/google/rtdn [post]
func RTDN(c *gin.Context) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable body"})
		return
	}

	notif, err := playbilling.DecodeRTDN(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification"})
		return
	}

	// Test pings and non-subscription notifications: just acknowledge.
	if notif.TestNotification != nil || notif.SubscriptionNotification == nil {
		c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
		return
	}
	sn := notif.SubscriptionNotification

	userID, err := billingServices.FindUserByGoogleToken(c.Request.Context(), sn.PurchaseToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if userID == nil {
		// Purchase never verified through us → nothing to update. Ack.
		c.JSON(http.StatusOK, gin.H{"message": "acknowledged (no matching user)"})
		return
	}

	// Re-verify against the Play Developer API — authoritative source of truth.
	purchase, err := playbilling.VerifySubscription(c.Request.Context(), sn.SubscriptionID, sn.PurchaseToken)
	if err != nil {
		logger.Error(c.Request.Context(), err, "google rtdn: purchase re-verification failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verification failed"})
		return
	}

	status := statusFrom(sn.NotificationType, purchase)
	record := billingModels.GooglePlayPurchase{
		PurchaseToken:        sn.PurchaseToken,
		ProductID:            ptr(sn.SubscriptionID),
		OrderID:              ptr(purchase.OrderID),
		ExpiryTime:           purchase.ExpiryTime(),
		AcknowledgementState: &purchase.AcknowledgementState,
	}
	if _, err := billingServices.ApplyGooglePurchase(c.Request.Context(), *userID, record, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
}

// statusFrom derives subscription status from the RTDN type, falling back to the
// verified expiry when the type is ambiguous/unknown.
func statusFrom(notificationType int, purchase *playbilling.SubscriptionPurchase) billingModels.Status {
	entitled, recognized := playbilling.StatusForNotificationType(notificationType)
	if recognized {
		switch notificationType {
		case playbilling.RTDNRevoked:
			return billingModels.StatusRevoked
		case playbilling.RTDNCanceled:
			return billingModels.StatusCanceled
		case playbilling.RTDNInGracePeriod, playbilling.RTDNOnHold:
			return billingModels.StatusGracePeriod
		case playbilling.RTDNExpired, playbilling.RTDNPaused:
			return billingModels.StatusExpired
		}
		if entitled {
			return billingModels.StatusActive
		}
		return billingModels.StatusExpired
	}
	if exp := purchase.ExpiryTime(); exp != nil && exp.Before(time.Now()) {
		return billingModels.StatusExpired
	}
	return billingModels.StatusActive
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Package appleIAPHandlers handles App Store Server Notifications V2 — Apple's
// webhook-equivalent for renewals, cancellations, refunds, billing retries, and
// grace periods. Payloads are JWS-signed (not shared-secret), verified against
// the x5c certificate chain in integrations/appstore.
package appleIAPHandlers

import (
	"net/http"
	"time"

	"dalista/dalista-be/api/server/logger"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"
	billingServices "dalista/dalista-be/app/services/billingServices"
	"dalista/dalista-be/integrations/appstore"

	"github.com/gin-gonic/gin"
)

// Notifications godoc
//
//	@Summary	App Store Server Notifications V2 webhook
//	@Description	Apple delivers a JWS `signedPayload`; verified against the x5c chain, then the subscription state is synced. Public endpoint (authenticity is the signature).
//	@Tags		billing-webhooks
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	responses.ErrorResponse	"acknowledged"
//	@Failure	400	{object}	responses.ErrorResponse
//	@Failure	401	{object}	responses.ErrorResponse	"Invalid signature"
//	@Router		/billing/apple/notifications [post]
func Notifications(c *gin.Context) {
	var body struct {
		SignedPayload string `json:"signedPayload"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.SignedPayload == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing signedPayload"})
		return
	}

	notif, err := appstore.ParseNotification(body.SignedPayload)
	if err != nil {
		// Signature/chain failure — reject so Apple retries and we alert on it.
		logger.Error(c.Request.Context(), err, "apple notification signature verification failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	txn, err := appstore.DecodeTransaction(notif.Data.SignedTransactionInfo)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid transaction signature"})
		return
	}

	userID, err := billingServices.FindUserByAppleOriginalTxn(c.Request.Context(), txn.OriginalTransactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if userID == nil {
		// Unknown transaction (e.g. a purchase never verified through us). Ack so
		// Apple stops retrying; nothing to update.
		c.JSON(http.StatusOK, gin.H{"message": "acknowledged (no matching user)"})
		return
	}

	status := statusFrom(notif, txn)
	receipt := billingModels.AppleIAPReceipt{
		OriginalTransactionID: txn.OriginalTransactionID,
		TransactionID:         ptr(txn.TransactionID),
		ProductID:             ptr(txn.ProductID),
		Environment:           ptr(txn.Environment),
		ExpiresAt:             txn.ExpiresAt(),
	}
	if _, err := billingServices.ApplyAppleReceipt(c.Request.Context(), *userID, receipt, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "acknowledged"})
}

// statusFrom maps an Apple notification + transaction to a subscription status.
func statusFrom(notif *appstore.Notification, txn *appstore.TransactionInfo) billingModels.Status {
	if txn.Revoked() {
		return billingModels.StatusRevoked
	}
	switch notif.NotificationType {
	case "REFUND", "REVOKE":
		return billingModels.StatusRevoked
	case "EXPIRED", "GRACE_PERIOD_EXPIRED":
		return billingModels.StatusExpired
	case "DID_RENEW", "SUBSCRIBED", "OFFER_REDEEMED", "RESUBSCRIBE":
		return billingModels.StatusActive
	case "DID_FAIL_TO_RENEW":
		if notif.Subtype == "GRACE_PERIOD" {
			return billingModels.StatusGracePeriod
		}
		return billingModels.StatusExpired
	default:
		if exp := txn.ExpiresAt(); exp != nil && exp.Before(time.Now()) {
			return billingModels.StatusExpired
		}
		return billingModels.StatusActive
	}
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

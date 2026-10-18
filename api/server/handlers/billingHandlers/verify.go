package billingHandlers

import (
	"errors"
	"net/http"
	"time"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"
	billingServices "dalista/dalista-be/app/services/billingServices"
	"dalista/dalista-be/app/utils"
	"dalista/dalista-be/integrations/appstore"
	"dalista/dalista-be/integrations/playbilling"

	"github.com/gin-gonic/gin"
)

// VerifyApple godoc
//
//	@Summary	Verify an Apple IAP purchase
//	@Description	Server-side validation of a StoreKit transaction via the App Store Server API, then activates the tier. The client never self-reports entitlement.
//	@Tags		billing
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.AppleVerifyForm	true	"Transaction id"
//	@Success	200		{object}	resources.SubscriptionResource
//	@Failure	402		{object}	responses.ErrorResponse	"Purchase not valid"
//	@Failure	503		{object}	responses.ErrorResponse	"Apple billing not configured"
//	@Router		/billing/apple/verify [post]
func VerifyApple(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, billingServices.PermReadSelfBilling) {
		return
	}
	var form forms.AppleVerifyForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}

	info, err := appstore.GetTransactionInfo(c.Request.Context(), form.TransactionID)
	if err != nil {
		if errors.Is(err, appstore.ErrNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Apple billing is not configured"})
			return
		}
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Could not validate purchase"})
		return
	}

	status := billingModels.StatusActive
	if info.Revoked() {
		status = billingModels.StatusRevoked
	} else if exp := info.ExpiresAt(); exp != nil && exp.Before(time.Now()) {
		status = billingModels.StatusExpired
	}

	receipt := billingModels.AppleIAPReceipt{
		OriginalTransactionID: info.OriginalTransactionID,
		TransactionID:         strPtr(info.TransactionID),
		ProductID:             strPtr(info.ProductID),
		Environment:           strPtr(info.Environment),
		ExpiresAt:             info.ExpiresAt(),
	}
	sub, err := billingServices.ApplyAppleReceipt(c.Request.Context(), uid, receipt, status)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, resources.SubscriptionFromModel(sub))
}

// VerifyGoogle godoc
//
//	@Summary	Verify a Google Play purchase
//	@Description	Server-side validation of a Play Billing purchase token via the Play Developer API, then activates the tier.
//	@Tags		billing
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.GoogleVerifyForm	true	"Purchase token + subscription id"
//	@Success	200		{object}	resources.SubscriptionResource
//	@Failure	402		{object}	responses.ErrorResponse	"Purchase not valid"
//	@Failure	503		{object}	responses.ErrorResponse	"Google billing not configured"
//	@Router		/billing/google/verify [post]
func VerifyGoogle(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, billingServices.PermReadSelfBilling) {
		return
	}
	var form forms.GoogleVerifyForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}

	purchase, err := playbilling.VerifySubscription(c.Request.Context(), form.SubscriptionID, form.PurchaseToken)
	if err != nil {
		if errors.Is(err, playbilling.ErrNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google billing is not configured"})
			return
		}
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Could not validate purchase"})
		return
	}

	status := billingModels.StatusActive
	if exp := purchase.ExpiryTime(); exp != nil && exp.Before(time.Now()) {
		status = billingModels.StatusExpired
	}
	if purchase.CancelReason != nil {
		status = billingModels.StatusCanceled
	}

	record := billingModels.GooglePlayPurchase{
		PurchaseToken:        form.PurchaseToken,
		ProductID:            strPtr(form.SubscriptionID),
		OrderID:              strPtr(purchase.OrderID),
		ExpiryTime:           purchase.ExpiryTime(),
		AcknowledgementState: &purchase.AcknowledgementState,
	}
	sub, err := billingServices.ApplyGooglePurchase(c.Request.Context(), uid, record, status)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, resources.SubscriptionFromModel(sub))
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

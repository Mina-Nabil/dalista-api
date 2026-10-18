package billingHandlers

import (
	"net/http"

	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	billingServices "dalista/dalista-be/app/services/billingServices"

	"github.com/gin-gonic/gin"
)

// GetSubscription godoc
//
//	@Summary	Get my subscription
//	@Description	Returns the caller's current subscription (or the implicit free tier).
//	@Tags		billing
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{object}	resources.SubscriptionResource
//	@Router		/billing/subscription [get]
func GetSubscription(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, billingServices.PermReadSelfBilling) {
		return
	}
	sub, err := billingServices.GetSubscription(c.Request.Context(), uid)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, resources.SubscriptionFromModel(sub))
}

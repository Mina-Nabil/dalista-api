package billingHandlers

import (
	"net/http"

	"dalista/dalista-be/api/server/middlewares"
	billingServices "dalista/dalista-be/app/services/billingServices"

	"github.com/gin-gonic/gin"
)

// SubscriptionSummary godoc
//
//	@Summary	Subscription summary (admin)
//	@Description	Counts of subscriptions by tier, billing rail, and status for the dashboard.
//	@Tags		billing-admin
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{object}	billingServices.SubscriptionSummary
//	@Failure	403	{object}	responses.ErrorResponse
//	@Router		/admin/subscriptions/summary [get]
func SubscriptionSummary(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		return
	}
	if !middlewares.AbortIfCant(c, billingServices.PermReadSubscriptions) {
		return
	}
	summary, err := billingServices.Summary(c.Request.Context())
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ListSubscriptions godoc
//
//	@Summary	List subscriptions (admin)
//	@Description	Subscriptions joined to owner contact fields, optionally filtered by tier and billing rail.
//	@Tags		billing-admin
//	@Security	BearerAuth
//	@Produce	json
//	@Param		tier		query	string	false	"Filter by tier (free/basic/unlimited)"
//	@Param		provider	query	string	false	"Filter by billing rail (apple_iap/google_play/stripe)"
//	@Success	200	{array}	billingServices.OversightRow
//	@Failure	403	{object}	responses.ErrorResponse
//	@Router		/admin/subscriptions [get]
func ListSubscriptions(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		return
	}
	if !middlewares.AbortIfCant(c, billingServices.PermReadSubscriptions) {
		return
	}
	rows, err := billingServices.List(c.Request.Context(), c.Query("tier"), c.Query("provider"))
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, rows)
}

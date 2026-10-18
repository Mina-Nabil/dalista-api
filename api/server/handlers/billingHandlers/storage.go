package billingHandlers

import (
	"net/http"

	"dalista/dalista-be/api/server/middlewares"
	billingServices "dalista/dalista-be/app/services/billingServices"

	"github.com/gin-gonic/gin"
)

// GetStorage godoc
//
//	@Summary	Get my storage usage vs. quota
//	@Tags		billing
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{object}	billingServices.StorageStatus
//	@Router		/billing/storage [get]
func GetStorage(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, billingServices.PermReadSelfBilling) {
		return
	}
	status, err := billingServices.GetStatus(c.Request.Context(), uid)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, status)
}

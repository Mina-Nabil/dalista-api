// Package notificationHandlers holds the HTTP handlers for notifications: push
// device-token registration, the in-app inbox, and the admin SMS send.
package notificationHandlers

import (
	"errors"
	"net/http"
	"strconv"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	notificationServices "dalista/dalista-be/app/services/notificationServices"
	"dalista/dalista-be/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	user, ok := middlewares.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return uuid.UUID{}, false
	}
	return user.ID, true
}

func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, notificationServices.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

// RegisterDevice godoc
//
//	@Summary	Register a push device token
//	@Description	Registers (or refreshes) the caller's FCM device token. Idempotent per token; a token re-registering under a new account is reassigned.
//	@Tags		notifications
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.RegisterDeviceForm	true	"Device token + platform"
//	@Success	200		{object}	resources.DeviceTokenResource
//	@Failure	400		{object}	responses.ErrorResponse
//	@Router		/devices [post]
func RegisterDevice(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, notificationServices.PermManageDevices) {
		return
	}
	var form forms.RegisterDeviceForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	dt, err := notificationServices.RegisterDevice(c.Request.Context(), uid, form.Token, form.Platform)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.DeviceTokenFromModel(dt))
}

// UnregisterDevice godoc
//
//	@Summary	Unregister a push device token
//	@Description	Removes one of the caller's device tokens (e.g. on logout). Idempotent.
//	@Tags		notifications
//	@Security	BearerAuth
//	@Accept		json
//	@Param		request	body	forms.UnregisterDeviceForm	true	"Device token"
//	@Success	204		"No Content"
//	@Failure	400		{object}	responses.ErrorResponse
//	@Router		/devices [delete]
func UnregisterDevice(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, notificationServices.PermManageDevices) {
		return
	}
	var form forms.UnregisterDeviceForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	if err := notificationServices.UnregisterDevice(c.Request.Context(), uid, form.Token); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListNotifications godoc
//
//	@Summary	List my notifications (inbox)
//	@Tags		notifications
//	@Security	BearerAuth
//	@Produce	json
//	@Param		limit	query	int	false	"Max to return (default 50, max 100)"
//	@Success	200	{array}	resources.NotificationResource
//	@Router		/notifications [get]
func ListNotifications(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, notificationServices.PermReadNotifications) {
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := notificationServices.ListForUser(c.Request.Context(), uid, limit)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.NotificationsFromModels(items))
}

// AdminSendSMS godoc
//
//	@Summary	Send an SMS to selected users (admin)
//	@Description	Admin-only. Sends a message to each selected user's phone via SMS Misr. Users without a phone are skipped.
//	@Tags		notifications
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.AdminSMSForm	true	"User IDs + message"
//	@Success	200		{object}	notificationServices.AdminSMSResult
//	@Failure	400		{object}	responses.ErrorResponse
//	@Failure	403		{object}	responses.ErrorResponse
//	@Router		/admin/sms [post]
func AdminSendSMS(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		return
	}
	if !middlewares.AbortIfCant(c, notificationServices.PermSendAdminSMS) {
		return
	}
	var form forms.AdminSMSForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	ids := make([]uuid.UUID, 0, len(form.UserIDs))
	for _, s := range form.UserIDs {
		if id, err := uuid.Parse(s); err == nil {
			ids = append(ids, id)
		}
	}
	res, err := notificationServices.SendAdminSMS(c.Request.Context(), ids, form.Message)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

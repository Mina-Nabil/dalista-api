// Package profileHandlers holds the handlers for the authenticated user's own
// profile (Module 2): GET and PUT /api/v1/profile.
package profileHandlers

import (
	"encoding/json"
	"net/http"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	authServices "dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/app/utils"

	"github.com/gin-gonic/gin"
)

// GetProfile godoc
//
//	@Summary      Get my profile
//	@Tags         profile
//	@Security     BearerAuth
//	@Produce      json
//	@Success      200  {object}  resources.ProfileResource
//	@Failure      401  {object}  responses.ErrorResponse
//	@Router       /profile [get]
func GetProfile(c *gin.Context) {
	user, ok := middlewares.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if !middlewares.AbortIfCant(c, authServices.PermReadProfile) {
		return
	}
	u, profile, err := authServices.GetProfile(c.Request.Context(), user.ID)
	if err != nil || u == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, resources.ProfileFromModel(u, profile))
}

// UpdateProfile godoc
//
//	@Summary      Update my profile
//	@Description  Partial update — only provided fields change.
//	@Tags         profile
//	@Security     BearerAuth
//	@Accept       json
//	@Produce      json
//	@Param        request  body      forms.UpdateProfileForm  true  "Fields to update"
//	@Success      200      {object}  resources.ProfileResource
//	@Failure      400      {object}  responses.ValidationErrorResponse
//	@Failure      401      {object}  responses.ErrorResponse
//	@Router       /profile [put]
func UpdateProfile(c *gin.Context) {
	user, ok := middlewares.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	if !middlewares.AbortIfCant(c, authServices.PermUpdateProfile) {
		return
	}

	var form forms.UpdateProfileForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}

	update := authServices.ProfileUpdate{
		DisplayName: form.DisplayName,
		PictureURL:  form.PictureURL,
	}
	if form.NotificationPrefs != nil {
		raw, err := json.Marshal(form.NotificationPrefs)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": []string{"notification_prefs must be a JSON object"}})
			return
		}
		s := string(raw)
		update.NotificationPrefs = &s
	}

	u, profile, err := authServices.UpdateProfile(c.Request.Context(), user.ID, update)
	if err != nil || u == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, resources.ProfileFromModel(u, profile))
}

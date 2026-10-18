// Package periodicHandlers holds the HTTP handlers for periodic lists: the
// need-to-buy toggle, the "create today's list" review + generation, and
// variant history.
package periodicHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	listServices "dalista/dalista-be/app/services/listServices"
	periodicServices "dalista/dalista-be/app/services/periodicServices"
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

func parseUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return uuid.UUID{}, false
	}
	return id, true
}

func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, periodicServices.ErrNotPeriodic):
		c.JSON(http.StatusBadRequest, gin.H{"error": "List is not periodic"})
	case errors.Is(err, periodicServices.ErrNotFound), errors.Is(err, listServices.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	case errors.Is(err, listServices.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

// SetNeedToBuy godoc
//
//	@Summary	Toggle an item's need-to-buy state (periodic lists)
//	@Description	Editor+ on a periodic list. Records a state transition and an auto system comment ("Marked as need to buy" / "We have enough").
//	@Tags		periodic
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"Item ID"
//	@Param		request	body		forms.NeedToBuyForm	true	"need_to_buy"
//	@Success	200		{object}	resources.ItemResource
//	@Failure	400		{object}	responses.ErrorResponse	"Not a periodic list"
//	@Failure	403		{object}	responses.ErrorResponse
//	@Router		/items/{id}/need-to-buy [put]
func SetNeedToBuy(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, periodicServices.PermManagePeriodic) {
		return
	}
	itemID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var form forms.NeedToBuyForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	item, err := periodicServices.ToggleNeedToBuy(c.Request.Context(), uid, itemID, *form.NeedToBuy)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ItemFromModel(item))
}

// TodayReview godoc
//
//	@Summary	Review items for "today's list"
//	@Description	Returns the periodic list's items currently marked need-to-buy — the candidate set to confirm/exclude before generating a variant.
//	@Tags		periodic
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	string	true	"Periodic list ID"
//	@Success	200	{array}	resources.ItemResource
//	@Failure	400	{object}	responses.ErrorResponse	"Not a periodic list"
//	@Router		/lists/{id}/today [get]
func TodayReview(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadList) {
		return
	}
	listID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	_, items, err := periodicServices.TodayReview(c.Request.Context(), uid, listID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ItemsFromModels(items))
}

// GenerateToday godoc
//
//	@Summary	Generate "today's list" (a dated variant)
//	@Description	Snapshots the current need-to-buy items (minus any excluded, which are marked "enough") into a new dated ListVariant.
//	@Tags		periodic
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"Periodic list ID"
//	@Param		request	body		forms.GenerateTodayForm	false	"Items to exclude"
//	@Success	201		{object}	resources.ListVariantResource
//	@Failure	400		{object}	responses.ErrorResponse	"Not a periodic list"
//	@Failure	403		{object}	responses.ErrorResponse
//	@Router		/lists/{id}/today [post]
func GenerateToday(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, periodicServices.PermManagePeriodic) {
		return
	}
	listID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var form forms.GenerateTodayForm
	if err := c.ShouldBindJSON(&form); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	excludes := make([]uuid.UUID, 0, len(form.ExcludeItemIDs))
	for _, s := range form.ExcludeItemIDs {
		if id, err := uuid.Parse(s); err == nil {
			excludes = append(excludes, id)
		}
	}
	variant, err := periodicServices.GenerateToday(c.Request.Context(), uid, listID, excludes)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.ListVariantFromModel(variant))
}

// ListVariants godoc
//
//	@Summary	List a periodic list's variants (history)
//	@Tags		periodic
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	string	true	"Periodic list ID"
//	@Success	200	{array}	resources.ListVariantResource
//	@Router		/lists/{id}/variants [get]
func ListVariants(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadList) {
		return
	}
	listID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	variants, err := periodicServices.ListVariants(c.Request.Context(), uid, listID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ListVariantsSummary(variants))
}

// GetVariant godoc
//
//	@Summary	Get a variant with its items
//	@Tags		periodic
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"Variant ID"
//	@Success	200	{object}	resources.ListVariantResource
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/variants/{id} [get]
func GetVariant(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadList) {
		return
	}
	variantID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	variant, err := periodicServices.GetVariant(c.Request.Context(), uid, variantID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ListVariantFromModel(variant))
}

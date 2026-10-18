package listHandlers

import (
	"net/http"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	listServices "dalista/dalista-be/app/services/listServices"
	"dalista/dalista-be/app/utils"

	"github.com/gin-gonic/gin"
)

// ListGroupedLists godoc
//
//	@Summary	List grouped lists
//	@Tags		grouped-lists
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	resources.GroupedListResource
//	@Router		/grouped-lists [get]
func ListGroupedLists(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadGroupedList) {
		return
	}
	gls, err := listServices.ListGroupedLists(c.Request.Context(), uid)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.GroupedListsFromModels(gls))
}

// CreateGroupedList godoc
//
//	@Summary	Create a grouped list
//	@Tags		grouped-lists
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.GroupedListForm	true	"Grouped list"
//	@Success	201		{object}	resources.GroupedListResource
//	@Router		/grouped-lists [post]
func CreateGroupedList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermCreateGroupedList) {
		return
	}
	var form forms.GroupedListForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	gl, err := listServices.CreateGroupedList(c.Request.Context(), uid, form.Name)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.GroupedListFromModel(gl))
}

// UpdateGroupedList godoc
//
//	@Summary	Rename a grouped list
//	@Tags		grouped-lists
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"Grouped list ID"
//	@Param		request	body		forms.GroupedListForm	true	"Grouped list"
//	@Success	200		{object}	resources.GroupedListResource
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/grouped-lists/{id} [put]
func UpdateGroupedList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermUpdateGroupedList) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	var form forms.GroupedListForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	gl, err := listServices.UpdateGroupedList(c.Request.Context(), uid, id, form.Name)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.GroupedListFromModel(gl))
}

// DeleteGroupedList godoc
//
//	@Summary	Delete a grouped list (nested lists become standalone)
//	@Tags		grouped-lists
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Grouped list ID"
//	@Success	204	"No Content"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/grouped-lists/{id} [delete]
func DeleteGroupedList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermDeleteGroupedList) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	if err := listServices.DeleteGroupedList(c.Request.Context(), uid, id); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

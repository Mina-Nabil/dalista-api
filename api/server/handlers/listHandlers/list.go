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

// ListLists godoc
//
//	@Summary	List lists (optionally filtered by grouped list)
//	@Tags		lists
//	@Security	BearerAuth
//	@Produce	json
//	@Param		grouped_list_id	query	string	false	"Filter by grouped list ID"
//	@Success	200	{array}	resources.ListResource
//	@Router		/lists [get]
func ListLists(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadList) {
		return
	}
	var groupedListID *string
	if q := c.Query("grouped_list_id"); q != "" {
		groupedListID = &q
	}
	glID, err := parseOptionalUUID(groupedListID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid grouped_list_id"})
		return
	}
	lists, err := listServices.ListLists(c.Request.Context(), uid, glID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ListsFromModels(lists))
}

// GetList godoc
//
//	@Summary	Get a list
//	@Tags		lists
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"List ID"
//	@Success	200	{object}	resources.ListResource
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/lists/{id} [get]
func GetList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadList) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	list, err := listServices.GetList(c.Request.Context(), uid, id)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ListFromModel(list))
}

// CreateList godoc
//
//	@Summary	Create a list (standalone or nested)
//	@Tags		lists
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.CreateListForm	true	"List"
//	@Success	201		{object}	resources.ListResource
//	@Failure	400		{object}	responses.ErrorResponse
//	@Router		/lists [post]
func CreateList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermCreateList) {
		return
	}
	var form forms.CreateListForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	glID, err := parseOptionalUUID(form.GroupedListID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid grouped_list_id"})
		return
	}
	catID, err := parseOptionalUUID(form.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	list, err := listServices.CreateList(c.Request.Context(), uid, listServices.ListInput{
		Name:          form.Name,
		GroupedListID: glID,
		CategoryID:    catID,
		IsPublic:      form.IsPublic,
		IsPeriodic:    form.IsPeriodic,
	})
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.ListFromModel(list))
}

// UpdateList godoc
//
//	@Summary	Update a list
//	@Tags		lists
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"List ID"
//	@Param		request	body		forms.UpdateListForm	true	"Fields to update"
//	@Success	200		{object}	resources.ListResource
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/lists/{id} [put]
func UpdateList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermUpdateList) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	var form forms.UpdateListForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	glID, err := parseOptionalUUID(form.GroupedListID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid grouped_list_id"})
		return
	}
	catID, err := parseOptionalUUID(form.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	in := listServices.ListInput{
		GroupedListID: glID,
		CategoryID:    catID,
		IsPublic:      form.IsPublic,
		IsPeriodic:    form.IsPeriodic,
	}
	if form.Name != nil {
		in.Name = *form.Name
	}
	list, err := listServices.UpdateList(c.Request.Context(), uid, id, in, form.ClearGroupedList, form.ClearCategory)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ListFromModel(list))
}

// DeleteList godoc
//
//	@Summary	Delete a list (and its items)
//	@Tags		lists
//	@Security	BearerAuth
//	@Param		id	path	string	true	"List ID"
//	@Success	204	"No Content"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/lists/{id} [delete]
func DeleteList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermDeleteList) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	if err := listServices.DeleteList(c.Request.Context(), uid, id); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

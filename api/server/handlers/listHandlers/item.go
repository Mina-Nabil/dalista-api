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

// ListItems godoc
//
//	@Summary	List items in a list
//	@Tags		items
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id		path	string	true	"List ID"
//	@Param		misc	query	bool	false	"Only MISC (uncategorized) items"
//	@Success	200	{array}	resources.ItemResource
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/lists/{id}/items [get]
func ListItems(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadItem) {
		return
	}
	listID, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	items, err := listServices.ListItems(c.Request.Context(), uid, listID, c.Query("misc") == "true")
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ItemsFromModels(items))
}

// CreateItem godoc
//
//	@Summary	Add an item to a list (omit category_id for MISC)
//	@Tags		items
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"List ID"
//	@Param		request	body		forms.CreateItemForm	true	"Item"
//	@Success	201		{object}	resources.ItemResource
//	@Failure	400		{object}	responses.ErrorResponse
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/lists/{id}/items [post]
func CreateItem(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermCreateItem) {
		return
	}
	listID, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	var form forms.CreateItemForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	catID, err := parseOptionalUUID(form.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	attrs, err := marshalAttributes(form.Attributes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attributes"})
		return
	}
	item, err := listServices.CreateItem(c.Request.Context(), uid, listID, listServices.ItemInput{
		Name:       form.Name,
		CategoryID: catID,
		Brand:      form.Brand,
		Price:      form.Price,
		Notes:      form.Notes,
		Attributes: attrs,
	})
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.ItemFromModel(item))
}

// GetItem godoc
//
//	@Summary	Get an item
//	@Tags		items
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"Item ID"
//	@Success	200	{object}	resources.ItemResource
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/items/{id} [get]
func GetItem(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadItem) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	item, err := listServices.GetItem(c.Request.Context(), uid, id)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ItemFromModel(item))
}

// UpdateItem godoc
//
//	@Summary	Update an item (clear_category moves it to MISC)
//	@Tags		items
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"Item ID"
//	@Param		request	body		forms.UpdateItemForm	true	"Fields to update"
//	@Success	200		{object}	resources.ItemResource
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/items/{id} [put]
func UpdateItem(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermUpdateItem) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	var form forms.UpdateItemForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	catID, err := parseOptionalUUID(form.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	attrs, err := marshalAttributes(form.Attributes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid attributes"})
		return
	}
	in := listServices.ItemInput{
		CategoryID: catID,
		Brand:      form.Brand,
		Price:      form.Price,
		Notes:      form.Notes,
		Attributes: attrs,
	}
	if form.Name != nil {
		in.Name = *form.Name
	}
	item, err := listServices.UpdateItem(c.Request.Context(), uid, id, in, form.ClearCategory)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ItemFromModel(item))
}

// DeleteItem godoc
//
//	@Summary	Delete an item
//	@Tags		items
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Item ID"
//	@Success	204	"No Content"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/items/{id} [delete]
func DeleteItem(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermDeleteItem) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	if err := listServices.DeleteItem(c.Request.Context(), uid, id); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

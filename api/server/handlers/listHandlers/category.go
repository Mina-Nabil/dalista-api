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

// ListCategories godoc
//
//	@Summary	List categories (own + system defaults)
//	@Tags		categories
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	resources.CategoryResource
//	@Router		/categories [get]
func ListCategories(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadCategory) {
		return
	}
	cats, err := listServices.ListCategories(c.Request.Context(), uid)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.CategoriesFromModels(cats))
}

// CreateCategory godoc
//
//	@Summary	Create a category
//	@Tags		categories
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.CategoryForm	true	"Category"
//	@Success	201		{object}	resources.CategoryResource
//	@Router		/categories [post]
func CreateCategory(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermCreateCategory) {
		return
	}
	var form forms.CategoryForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	cat, err := listServices.CreateCategory(c.Request.Context(), uid, form.Name)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.CategoryFromModel(cat))
}

// UpdateCategory godoc
//
//	@Summary	Rename a category
//	@Tags		categories
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"Category ID"
//	@Param		request	body		forms.CategoryForm	true	"Category"
//	@Success	200		{object}	resources.CategoryResource
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermUpdateCategory) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	var form forms.CategoryForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	cat, err := listServices.UpdateCategory(c.Request.Context(), uid, id, form.Name)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.CategoryFromModel(cat))
}

// DeleteCategory godoc
//
//	@Summary	Delete a category
//	@Tags		categories
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Category ID"
//	@Success	204	"No Content"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermDeleteCategory) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	if err := listServices.DeleteCategory(c.Request.Context(), uid, id); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

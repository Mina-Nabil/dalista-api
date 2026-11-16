// Package comparisonHandlers holds the HTTP handlers for item comparison:
// creating a saved comparison, viewing it (owner or via a share token) with
// per-item access gating, and generating a share token.
package comparisonHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	comparisonServices "dalista/dalista-be/app/services/comparisonServices"
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

func parsePathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return uuid.UUID{}, false
	}
	return id, true
}

func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, comparisonServices.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Need at least two items you can access"})
	case errors.Is(err, comparisonServices.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	case errors.Is(err, comparisonServices.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

// Create godoc
//
//	@Summary	Create a comparison session
//	@Description	Persists a comparison from a manually selected set of item IDs (own, shared, or public catalog). At least two accessible items required.
//	@Tags		comparisons
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		request	body		forms.CreateComparisonForm	true	"Item IDs to compare"
//	@Success	201		{object}	resources.ComparisonResource
//	@Failure	400		{object}	responses.ErrorResponse	"Fewer than two accessible items"
//	@Router		/comparisons [post]
func Create(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, comparisonServices.PermCreateComparison) {
		return
	}
	var form forms.CreateComparisonForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	ids := make([]uuid.UUID, 0, len(form.ItemIDs))
	for _, s := range form.ItemIDs {
		if id, err := uuid.Parse(s); err == nil {
			ids = append(ids, id)
		}
	}
	view, err := comparisonServices.Create(c.Request.Context(), uid, ids)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.ComparisonFromView(view))
}

// Get godoc
//
//	@Summary	Get a comparison (owner)
//	@Description	Returns a comparison you own, with its items. Items you can no longer access are withheld (see the withheld count).
//	@Tags		comparisons
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"Comparison ID"
//	@Success	200	{object}	resources.ComparisonResource
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/comparisons/{id} [get]
func Get(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, comparisonServices.PermReadComparison) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	view, err := comparisonServices.Get(c.Request.Context(), uid, id)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ComparisonFromView(view))
}

// Share godoc
//
//	@Summary	Share a comparison
//	@Description	Generates (or returns the existing) share token for a comparison you own. The token is the shareable reference; it grants no new item access — recipients only see items they can already access.
//	@Tags		comparisons
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"Comparison ID"
//	@Success	200	{object}	resources.ComparisonResource
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/comparisons/{id}/share [post]
func Share(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, comparisonServices.PermShareComparison) {
		return
	}
	id, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	session, err := comparisonServices.Share(c.Request.Context(), uid, id)
	if err != nil {
		mapErr(c, err)
		return
	}
	// Return the full view (owner sees all their items) with the token attached.
	view, err := comparisonServices.Get(c.Request.Context(), uid, session.ID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ComparisonFromView(view))
}

// GetShared godoc
//
//	@Summary	View a shared comparison
//	@Description	Returns a shared comparison by its token, showing only the items the viewer can access (public catalog or lists they can read). Others are withheld.
//	@Tags		comparisons
//	@Security	BearerAuth
//	@Produce	json
//	@Param		token	path		string	true	"Share token"
//	@Success	200		{object}	resources.ComparisonResource
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/shared-comparisons/{token} [get]
func GetShared(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, comparisonServices.PermReadComparison) {
		return
	}
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	view, err := comparisonServices.GetByToken(c.Request.Context(), uid, token)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ComparisonFromView(view))
}

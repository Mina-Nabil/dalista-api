// Package publicListHandlers holds the HTTP handlers for public lists &
// discovery: the star-sorted discovery feed, star/unstar, and the admin
// engagement analytics. (The "make a list public" permission gate lives on the
// list create/update handlers in listHandlers.)
package publicListHandlers

import (
	"errors"
	"net/http"
	"strconv"

	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	publicListServices "dalista/dalista-be/app/services/publicListServices"

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
	case errors.Is(err, publicListServices.ErrNotPublic):
		c.JSON(http.StatusBadRequest, gin.H{"error": "List is not public"})
	case errors.Is(err, publicListServices.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

// Discover godoc
//
//	@Summary	Discover public lists (sorted by stars)
//	@Description	Paginated feed of public lists, most-starred first, optionally filtered by category. Each item includes whether you've starred it.
//	@Tags		public-lists
//	@Security	BearerAuth
//	@Produce	json
//	@Param		category_id	query		string	false	"Filter by category ID"
//	@Param		limit		query		int		false	"Page size (default 20, max 100)"
//	@Param		offset		query		int		false	"Offset for pagination"
//	@Success	200			{object}	resources.DiscoveryFeedResource
//	@Router		/public-lists [get]
func Discover(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, publicListServices.PermDiscover) {
		return
	}

	var categoryID *uuid.UUID
	if q := c.Query("category_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		categoryID = &id
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	items, total, err := publicListServices.DiscoverFeed(c.Request.Context(), uid, categoryID, limit, offset)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.DiscoveryFeedFromModels(items, publicListServices.ClampLimit(limit), offset, total))
}

// Star godoc
//
//	@Summary	Star a public list
//	@Description	Stars a public list for the caller (idempotent). Returns the current star count.
//	@Tags		public-lists
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"List ID"
//	@Success	200	{object}	resources.StarResource
//	@Failure	400	{object}	responses.ErrorResponse	"List is not public"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/public-lists/{id}/star [post]
func Star(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, publicListServices.PermStar) {
		return
	}
	listID, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	state, err := publicListServices.Star(c.Request.Context(), uid, listID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.StarResourceFrom(listID.String(), state))
}

// Unstar godoc
//
//	@Summary	Unstar a public list
//	@Description	Removes the caller's star from a public list (idempotent). Returns the current star count.
//	@Tags		public-lists
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"List ID"
//	@Success	200	{object}	resources.StarResource
//	@Failure	400	{object}	responses.ErrorResponse	"List is not public"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/public-lists/{id}/star [delete]
func Unstar(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, publicListServices.PermStar) {
		return
	}
	listID, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	state, err := publicListServices.Unstar(c.Request.Context(), uid, listID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.StarResourceFrom(listID.String(), state))
}

// Engagement godoc
//
//	@Summary	Public list engagement analytics (admin)
//	@Description	Totals (public lists, stars) plus the top public lists by star count, for the dashboard.
//	@Tags		public-lists
//	@Security	BearerAuth
//	@Produce	json
//	@Param		top	query		int	false	"Number of top lists to return (default 10)"
//	@Success	200	{object}	publicListServices.Engagement
//	@Failure	403	{object}	responses.ErrorResponse
//	@Router		/admin/public-lists/engagement [get]
func Engagement(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		return
	}
	if !middlewares.AbortIfCant(c, publicListServices.PermReadEngagement) {
		return
	}
	top, _ := strconv.Atoi(c.Query("top"))
	stats, err := publicListServices.EngagementStats(c.Request.Context(), top)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Package activityHandlers holds the HTTP handlers for per-item comments and
// the access-scoped activity timeline.
package activityHandlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	activityServices "dalista/dalista-be/app/services/activityServices"
	listServices "dalista/dalista-be/app/services/listServices"
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
	case errors.Is(err, activityServices.ErrNotFound), errors.Is(err, listServices.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	case errors.Is(err, activityServices.ErrForbidden), errors.Is(err, listServices.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

// CreateComment godoc
//
//	@Summary	Comment on an item
//	@Description	Any list member may comment, including viewers. Records a comment_added activity event.
//	@Tags		comments
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"Item ID"
//	@Param		request	body		forms.CommentForm	true	"Comment body"
//	@Success	201		{object}	resources.CommentResource
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/items/{id}/comments [post]
func CreateComment(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, activityServices.PermCreateComment) {
		return
	}
	itemID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var form forms.CommentForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	// GetItem enforces read access (owner/editor/viewer); viewers may comment.
	item, err := listServices.GetItem(c.Request.Context(), uid, itemID)
	if err != nil {
		mapErr(c, err)
		return
	}
	comment, err := activityServices.CreateComment(c.Request.Context(), uid, item.ID, item.ListID, form.Body)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.CommentFromModel(comment))
}

// ListComments godoc
//
//	@Summary	List an item's comments
//	@Tags		comments
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	string	true	"Item ID"
//	@Success	200	{array}	resources.CommentResource
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/items/{id}/comments [get]
func ListComments(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, activityServices.PermReadComment) {
		return
	}
	itemID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if _, err := listServices.GetItem(c.Request.Context(), uid, itemID); err != nil {
		mapErr(c, err)
		return
	}
	comments, err := activityServices.ListComments(c.Request.Context(), itemID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.CommentsFromModels(comments))
}

// UpdateComment godoc
//
//	@Summary	Edit a comment (author only)
//	@Tags		comments
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"Comment ID"
//	@Param		request	body		forms.CommentForm	true	"Comment body"
//	@Success	200		{object}	resources.CommentResource
//	@Failure	403		{object}	responses.ErrorResponse
//	@Router		/comments/{id} [put]
func UpdateComment(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, activityServices.PermUpdateComment) {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var form forms.CommentForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	comment, err := activityServices.UpdateComment(c.Request.Context(), uid, id, form.Body)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.CommentFromModel(comment))
}

// DeleteComment godoc
//
//	@Summary	Delete a comment (author or list owner)
//	@Tags		comments
//	@Security	BearerAuth
//	@Param		id	path	string	true	"Comment ID"
//	@Success	204	"No Content"
//	@Failure	403	{object}	responses.ErrorResponse
//	@Router		/comments/{id} [delete]
func DeleteComment(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, activityServices.PermDeleteComment) {
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := activityServices.DeleteComment(c.Request.Context(), uid, id); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Timeline godoc
//
//	@Summary	Activity timeline (access-scoped)
//	@Description	Chronological feed of activity across the lists you can access, newest first. Paginate with `before` (RFC3339).
//	@Tags		activity
//	@Security	BearerAuth
//	@Produce	json
//	@Param		limit	query	int		false	"Max events (1-100, default 50)"
//	@Param		before	query	string	false	"Return events older than this RFC3339 timestamp"
//	@Success	200	{array}	resources.ActivityEventResource
//	@Router		/timeline [get]
func Timeline(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, activityServices.PermReadActivity) {
		return
	}
	limit := 0
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	var before *time.Time
	if b := c.Query("before"); b != "" {
		if t, err := time.Parse(time.RFC3339, b); err == nil {
			before = &t
		}
	}
	events, err := activityServices.Timeline(c.Request.Context(), uid, limit, before)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.ActivityEventsFromModels(events))
}

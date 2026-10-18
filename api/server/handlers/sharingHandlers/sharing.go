// Package sharingHandlers holds the HTTP handlers for list sharing & invites:
// share-by-phone, role change, revoke, cancel invite, collaborator listing, and
// shared-with-me.
package sharingHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/forms"
	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	listServices "dalista/dalista-be/app/services/listServices"
	sharingServices "dalista/dalista-be/app/services/sharingServices"
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

// mapErr maps sharing-service errors to HTTP responses.
func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sharingServices.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	case errors.Is(err, sharingServices.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	case errors.Is(err, sharingServices.ErrCapReached):
		c.JSON(http.StatusConflict, gin.H{"error": "Per-list user cap reached for your plan"})
	case errors.Is(err, sharingServices.ErrSelfShare):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot share a list with yourself"})
	case errors.Is(err, sharingServices.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

// ShareList godoc
//
//	@Summary	Share a list by phone (owner only)
//	@Description	Creates an immediate share if the phone belongs to a registered user, else a pending invite auto-attached when that number is verified. Enforces the owner's tier per-list user cap.
//	@Tags		sharing
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string				true	"List ID"
//	@Param		request	body		forms.ShareListForm	true	"Phone + role"
//	@Success	201		{object}	resources.ShareResultResource
//	@Failure	400		{object}	responses.ErrorResponse
//	@Failure	404		{object}	responses.ErrorResponse
//	@Failure	409		{object}	responses.ErrorResponse	"Per-list cap reached"
//	@Router		/lists/{id}/shares [post]
func ShareList(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, sharingServices.PermShareList) {
		return
	}
	listID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var form forms.ShareListForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	res, err := sharingServices.ShareListByPhone(c.Request.Context(), uid, listID, form.Phone, form.Role)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.ShareResultFromModel(res))
}

// ChangeRole godoc
//
//	@Summary	Change a collaborator's role (owner only)
//	@Tags		sharing
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	string				true	"List ID"
//	@Param		userId	path	string				true	"Collaborator user ID"
//	@Param		request	body	forms.ChangeRoleForm	true	"Role"
//	@Success	204	"No Content"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/lists/{id}/shares/{userId} [put]
func ChangeRole(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, sharingServices.PermShareList) {
		return
	}
	listID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	targetID, ok := parseUUID(c, "userId")
	if !ok {
		return
	}
	var form forms.ChangeRoleForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": utils.FormatValidationErrors(err)})
		return
	}
	if err := sharingServices.ChangeRole(c.Request.Context(), uid, listID, targetID, form.Role); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RevokeShare godoc
//
//	@Summary	Revoke a collaborator (owner only)
//	@Tags		sharing
//	@Security	BearerAuth
//	@Param		id		path	string	true	"List ID"
//	@Param		userId	path	string	true	"Collaborator user ID"
//	@Success	204	"No Content"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/lists/{id}/shares/{userId} [delete]
func RevokeShare(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, sharingServices.PermShareList) {
		return
	}
	listID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	targetID, ok := parseUUID(c, "userId")
	if !ok {
		return
	}
	if err := sharingServices.Revoke(c.Request.Context(), uid, listID, targetID); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// CancelInvite godoc
//
//	@Summary	Cancel a pending invite (owner only)
//	@Tags		sharing
//	@Security	BearerAuth
//	@Param		id			path	string	true	"List ID"
//	@Param		inviteId	path	string	true	"Pending invite ID"
//	@Success	204	"No Content"
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/lists/{id}/invites/{inviteId} [delete]
func CancelInvite(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, sharingServices.PermShareList) {
		return
	}
	listID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	inviteID, ok := parseUUID(c, "inviteId")
	if !ok {
		return
	}
	if err := sharingServices.CancelInvite(c.Request.Context(), uid, listID, inviteID); err != nil {
		mapErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListCollaborators godoc
//
//	@Summary	List a list's collaborators (+ invites, for the owner)
//	@Tags		sharing
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path		string	true	"List ID"
//	@Success	200	{object}	sharingServices.Collaborators
//	@Failure	404	{object}	responses.ErrorResponse
//	@Router		/lists/{id}/collaborators [get]
func ListCollaborators(c *gin.Context) {
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
	out, err := sharingServices.ListCollaborators(c.Request.Context(), uid, listID)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// SharedWithMe godoc
//
//	@Summary	Lists shared with me
//	@Tags		sharing
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	resources.SharedListResource
//	@Router		/shared-lists [get]
func SharedWithMe(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermReadList) {
		return
	}
	rows, err := sharingServices.ListSharedWithMe(c.Request.Context(), uid)
	if err != nil {
		mapErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resources.SharedListsFromModels(rows))
}

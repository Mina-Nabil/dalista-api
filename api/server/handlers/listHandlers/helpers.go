// Package listHandlers holds the HTTP handlers for the core list domain:
// categories, grouped lists, lists, items, and the category-interest export.
package listHandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"dalista/dalista-be/api/server/middlewares"
	listServices "dalista/dalista-be/app/services/listServices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// currentUserID returns the authenticated user's id, or false if absent.
func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	user, ok := middlewares.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return uuid.UUID{}, false
	}
	return user.ID, true
}

// parsePathUUID parses a UUID path param, writing 404 and returning false on failure.
func parsePathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return uuid.UUID{}, false
	}
	return id, true
}

// parseOptionalUUID parses an optional *string UUID (from a form) into a *uuid.UUID.
func parseOptionalUUID(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// mapErr maps list-service errors to HTTP responses.
func mapErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, listServices.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	case errors.Is(err, listServices.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	case errors.Is(err, listServices.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reference (grouped list or category)"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}

// marshalAttributes turns a form's attributes map into a JSON string pointer.
func marshalAttributes(attrs map[string]interface{}) (*string, error) {
	if attrs == nil {
		return nil, nil
	}
	raw, err := json.Marshal(attrs)
	if err != nil {
		return nil, err
	}
	s := string(raw)
	return &s, nil
}

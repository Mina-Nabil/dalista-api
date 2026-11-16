package listHandlers

import (
	"errors"
	"net/http"

	"dalista/dalista-be/api/resources"
	"dalista/dalista-be/api/server/middlewares"
	billingServices "dalista/dalista-be/app/services/billingServices"
	listServices "dalista/dalista-be/app/services/listServices"

	"github.com/gin-gonic/gin"
)

// AddItemImage godoc
//
//	@Summary	Upload a photo for an item (multipart)
//	@Description	Editor+. Field name `image`. Max 10 MB, up to 6 images per item; types JPEG/PNG/WebP/HEIC/GIF. Size counts against the storage quota.
//	@Tags		items
//	@Security	BearerAuth
//	@Accept		mpfd
//	@Produce	json
//	@Param		id		path		string	true	"Item ID"
//	@Param		image	formData	file	true	"Image file"
//	@Success	201		{object}	resources.ItemImageResource
//	@Failure	400		{object}	responses.ErrorResponse
//	@Failure	403		{object}	responses.ErrorResponse
//	@Failure	413		{object}	responses.ErrorResponse
//	@Failure	415		{object}	responses.ErrorResponse
//	@Router		/items/{id}/images [post]
func AddItemImage(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermUpdateItem) {
		return
	}
	itemID, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image file (field 'image')"})
		return
	}
	// Reject oversize before opening to avoid buffering large files.
	if fileHeader.Size > listServices.MaxImageBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image exceeds 10 MB"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read image"})
		return
	}
	defer f.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	img, err := listServices.AddItemImage(c.Request.Context(), uid, itemID, contentType, fileHeader.Size, f)
	if err != nil {
		mapImageErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resources.ItemImageFromModel(img))
}

// DeleteItemImage godoc
//
//	@Summary	Delete an item photo
//	@Description	Editor+. Removes the stored object and releases the storage quota.
//	@Tags		items
//	@Security	BearerAuth
//	@Param		id		path	string	true	"Item ID"
//	@Param		imageId	path	string	true	"Image ID"
//	@Success	204		"No Content"
//	@Failure	403		{object}	responses.ErrorResponse
//	@Failure	404		{object}	responses.ErrorResponse
//	@Router		/items/{id}/images/{imageId} [delete]
func DeleteItemImage(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermUpdateItem) {
		return
	}
	itemID, ok := parsePathUUID(c, "id")
	if !ok {
		return
	}
	imageID, ok := parsePathUUID(c, "imageId")
	if !ok {
		return
	}
	if err := listServices.DeleteItemImage(c.Request.Context(), uid, itemID, imageID); err != nil {
		mapImageErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// mapImageErr maps image-specific service errors to HTTP responses, falling back
// to the shared mapErr for the common cases.
func mapImageErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, listServices.ErrUnsupportedMediaType):
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "unsupported image type (allowed: JPEG, PNG, WebP, HEIC, GIF)"})
	case errors.Is(err, listServices.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image exceeds 10 MB"})
	case errors.Is(err, listServices.ErrTooManyImages):
		c.JSON(http.StatusConflict, gin.H{"error": "image limit reached (max 6 per item)"})
	case errors.Is(err, billingServices.ErrQuotaExceeded):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "storage quota exceeded"})
	default:
		mapErr(c, err)
	}
}

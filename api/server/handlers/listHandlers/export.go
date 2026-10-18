package listHandlers

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"dalista/dalista-be/api/server/middlewares"
	listServices "dalista/dalista-be/app/services/listServices"

	"github.com/gin-gonic/gin"
)

// CategoryInterestExport godoc
//
//	@Summary	Export category interest as CSV (admin)
//	@Description	Per-user counts of owned items/lists by category. First cut — Phase 7 adds starred public lists.
//	@Tags		exports
//	@Security	BearerAuth
//	@Produce	text/csv
//	@Success	200	{string}	string	"CSV file"
//	@Failure	403	{object}	responses.ErrorResponse
//	@Router		/exports/category-interest [get]
func CategoryInterestExport(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		return
	}
	if !middlewares.AbortIfCant(c, listServices.PermExportCategoryInterest) {
		return
	}

	rows, err := listServices.CategoryInterest(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="category-interest.csv"`)

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"user_id", "email", "phone", "category", "item_count", "list_count"})
	for _, r := range rows {
		_ = w.Write([]string{
			r.UserID, r.Email, r.Phone, r.Category,
			strconv.Itoa(r.ItemCount), strconv.Itoa(r.ListCount),
		})
	}
	w.Flush()
}

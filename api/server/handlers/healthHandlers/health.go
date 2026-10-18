// Package healthHandlers exposes liveness/readiness probes for the API.
package healthHandlers

import (
	"net/http"

	"dalista/dalista-be/app/repository"

	"github.com/gin-gonic/gin"
)

// HealthResponse is the body returned by GET /api/v1/health.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
	DB     string `json:"db" example:"up"`
}

// Health reports API readiness, including Postgres connectivity.
//
// @Summary      Health check
// @Description  Returns 200 when the API and its database are reachable, 503 otherwise.
// @Tags         health
// @Produce      json
// @Success      200  {object}  healthHandlers.HealthResponse
// @Failure      503  {object}  healthHandlers.HealthResponse
// @Router       /health [get]
func Health(c *gin.Context) {
	dbStatus := "up"
	httpStatus := http.StatusOK

	if err := repository.Ping(); err != nil {
		dbStatus = "down"
		httpStatus = http.StatusServiceUnavailable
	}

	status := "ok"
	if httpStatus != http.StatusOK {
		status = "degraded"
	}

	c.JSON(httpStatus, HealthResponse{Status: status, DB: dbStatus})
}

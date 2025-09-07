package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"gorm.io/gorm"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	DB *gorm.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		DB: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "healthy",
	})
}

// ReadinessResponse represents the readiness check response structure
type ReadinessResponse struct {
	Status   string `json:"status" example:"ready"`
	Database string `json:"database,omitempty" example:"ok"`
}

// ReadinessCheck godoc
// @Summary Readiness check endpoint
// @Description Checks if the service is ready to handle requests by verifying database connectivity
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} ReadinessResponse
// @Failure 503 {object} ReadinessResponse
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	details := make(map[string]string)
	status := http.StatusOK

	// Check database connection
	db, err := h.DB.DB()
	if err != nil {
		details["database"] = "error: " + err.Error()
		status = http.StatusServiceUnavailable
	} else if err := db.Ping(); err != nil {
		details["database"] = "unhealthy: " + err.Error()
		status = http.StatusServiceUnavailable
	} else {
		details["database"] = "ok"
	}

	response := HealthResponse{
		Status:  map[bool]string{true: "ready", false: "not ready"}[status == http.StatusOK],
		Details: details,
	}

	c.JSON(status, response)
}

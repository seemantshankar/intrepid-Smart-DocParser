package router

import (
	"net/http"

	"contract-analysis-service/internal/handlers"
	"github.com/gin-gonic/gin"
)

type Router struct {
	r *gin.Engine
	h *handlers.HealthHandler
}

func NewRouter(healthHandler *handlers.HealthHandler) *Router {
	r := &Router{
		r: gin.Default(),
		h: healthHandler,
	}

	r.setupRoutes()

	return r
}

func (r *Router) setupRoutes() {
	// Health check routes
	r.r.GET("/health", r.h.HealthCheck)
	r.r.GET("/ready", r.h.ReadinessCheck)
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.r.ServeHTTP(w, req)
}

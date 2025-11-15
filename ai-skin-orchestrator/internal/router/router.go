package router

import (
	"github.com/aibanking/ai-skin-orchestrator/internal/controller"
	"github.com/aibanking/ai-skin-orchestrator/internal/middleware"
	"github.com/gorilla/mux"
)

// Router sets up all routes
type Router struct {
	orchestratorController *controller.OrchestratorController
	rateLimiter            *middleware.RateLimiter
}

// NewRouter creates a new router instance
func NewRouter(
	orchestratorController *controller.OrchestratorController,
	rateLimiter *middleware.RateLimiter,
) *Router {
	return &Router{
		orchestratorController: orchestratorController,
		rateLimiter:            rateLimiter,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Health check (no auth required)
	router.HandleFunc("/health", r.orchestratorController.HealthCheck).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/process", r.orchestratorController.ProcessRequest).Methods("POST")

	// Apply middleware (CORS first)
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.AuthMiddleware)
	router.Use(r.rateLimiter.RateLimitMiddleware)

	return router
}


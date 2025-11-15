package router

import (
	"github.com/aibanking/agent-mesh/internal/controller"
	"github.com/aibanking/agent-mesh/internal/middleware"
	"github.com/gorilla/mux"
)

// Router sets up all routes
type Router struct {
	agentController *controller.AgentController
	rateLimiter     *middleware.RateLimiter
}

// NewRouter creates a new router instance
func NewRouter(
	agentController *controller.AgentController,
	rateLimiter *middleware.RateLimiter,
) *Router {
	return &Router{
		agentController: agentController,
		rateLimiter:     rateLimiter,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Health check (no auth required)
	router.HandleFunc("/health", r.agentController.HealthCheck).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/process", r.agentController.ProcessRequest).Methods("POST")

	// Apply middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.AuthMiddleware)
	router.Use(r.rateLimiter.RateLimitMiddleware)

	return router
}


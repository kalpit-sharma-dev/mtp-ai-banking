package router

import (
	"github.com/aibanking/banking-integrations/internal/controller"
	"github.com/aibanking/banking-integrations/internal/middleware"
	"github.com/gorilla/mux"
)

// Router sets up all routes
type Router struct {
	bankingController *controller.BankingController
	rateLimiter       *middleware.RateLimiter
}

// NewRouter creates a new router instance
func NewRouter(
	bankingController *controller.BankingController,
	rateLimiter *middleware.RateLimiter,
) *Router {
	return &Router{
		bankingController: bankingController,
		rateLimiter:       rateLimiter,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Health check (no auth required)
	router.HandleFunc("/health", r.bankingController.HealthCheck).Methods("GET")

	// Banking API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/balance", r.bankingController.GetBalance).Methods("POST")
	api.HandleFunc("/transfer", r.bankingController.TransferFunds).Methods("POST")
	api.HandleFunc("/statement", r.bankingController.GetStatement).Methods("POST")
	api.HandleFunc("/beneficiary", r.bankingController.AddBeneficiary).Methods("POST")

	// DWH routes
	api.HandleFunc("/dwh/query", r.bankingController.QueryDWH).Methods("POST")
	api.HandleFunc("/dwh/history/{userID}", r.bankingController.GetTransactionHistory).Methods("GET")

	// Apply middleware (CORS first)
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.AuthMiddleware)
	router.Use(r.rateLimiter.RateLimitMiddleware)

	return router
}


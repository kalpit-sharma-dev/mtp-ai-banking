package router

import (
	"net/http"

	"github.com/aibanking/mcp-server/internal/controller"
	"github.com/aibanking/mcp-server/internal/middleware"
	"github.com/gorilla/mux"
)

// Router sets up all routes
type Router struct {
	taskController    *controller.TaskController
	agentController   *controller.AgentController
	sessionController *controller.SessionController
	ruleController    *controller.RuleController
	rateLimiter       *middleware.RateLimiter
}

// NewRouter creates a new router instance
func NewRouter(
	taskController *controller.TaskController,
	agentController *controller.AgentController,
	sessionController *controller.SessionController,
	ruleController *controller.RuleController,
	rateLimiter *middleware.RateLimiter,
) *Router {
	return &Router{
		taskController:    taskController,
		agentController:   agentController,
		sessionController: sessionController,
		ruleController:    ruleController,
		rateLimiter:       rateLimiter,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Health check endpoints
	router.HandleFunc("/health", r.healthCheck).Methods("GET")
	router.HandleFunc("/ready", r.readyCheck).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Task routes
	api.HandleFunc("/submit-task", r.taskController.SubmitTask).Methods("POST")
	api.HandleFunc("/get-result/{taskID}", r.taskController.GetTaskResult).Methods("GET")

	// Agent routes
	api.HandleFunc("/register-agent", r.agentController.RegisterAgent).Methods("POST")
	api.HandleFunc("/agent/{agentID}", r.agentController.GetAgent).Methods("GET")
	api.HandleFunc("/agents", r.agentController.GetAllAgents).Methods("GET")

	// Session routes
	api.HandleFunc("/get-session/{sessionID}", r.sessionController.GetSession).Methods("GET")
	api.HandleFunc("/create-session", r.sessionController.CreateSession).Methods("POST")

	// Rule routes
	api.HandleFunc("/rules/upload", r.ruleController.UploadRules).Methods("POST")
	api.HandleFunc("/rules", r.ruleController.GetRules).Methods("GET")

	// Apply middleware (CORS first)
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.AuthMiddleware)
	router.Use(r.rateLimiter.RateLimitMiddleware)

	return router
}

// healthCheck returns server health status
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// readyCheck returns server readiness status
func (r *Router) readyCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}


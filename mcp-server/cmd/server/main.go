package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aibanking/mcp-server/internal/config"
	"github.com/aibanking/mcp-server/internal/controller"
	"github.com/aibanking/mcp-server/internal/middleware"
	"github.com/aibanking/mcp-server/internal/model"
	"github.com/aibanking/mcp-server/internal/router"
	"github.com/aibanking/mcp-server/internal/service"
	"github.com/aibanking/mcp-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	utils.InitLogger(cfg.Logging.Level, cfg.Logging.Format)

	log.Info().Msg("Starting MCP Server for AI Banking Platform")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Redis, continuing without persistence")
	} else {
		log.Info().Msg("Connected to Redis")
	}

	// Initialize services
	sessionManager := service.NewSessionManager(redisClient)
	taskManager := service.NewTaskManager(redisClient)
	agentRegistry := service.NewAgentRegistry(redisClient)
	ruleEngine := service.NewRuleEngine()
	contextRouter := service.NewContextRouter(agentRegistry, ruleEngine)
	orchestrator := service.NewOrchestrator(sessionManager, taskManager, agentRegistry, contextRouter)

	// Initialize controllers
	taskController := controller.NewTaskController(orchestrator, taskManager)
	agentController := controller.NewAgentController(agentRegistry)
	sessionController := controller.NewSessionController(sessionManager)
	ruleController := controller.NewRuleController(ruleEngine)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Initialize router
	appRouter := router.NewRouter(
		taskController,
		agentController,
		sessionController,
		ruleController,
		rateLimiter,
	)

	// Setup routes
	r := appRouter.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("address", server.Addr).
			Msg("MCP Server started")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Register default agents (for testing/demo)
	registerDefaultAgents(ctx, agentRegistry)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

// registerDefaultAgents registers mock agents for demonstration
func registerDefaultAgents(ctx context.Context, registry *service.AgentRegistry) {
	defaultAgents := []struct {
		name         string
		agentType    string
		endpoint     string
		capabilities []string
	}{
		{
			name:         "Banking Agent",
			agentType:    "BANKING",
			endpoint:     "http://localhost:8001",
			capabilities: []string{"CHECK_BALANCE", "GET_STATEMENT", "FUND_TRANSFER"},
		},
		{
			name:         "Fraud Detection Agent",
			agentType:    "FRAUD",
			endpoint:     "http://localhost:8002",
			capabilities: []string{"FRAUD_CHECK", "RISK_ASSESSMENT"},
		},
		{
			name:         "Guardrail Agent",
			agentType:    "GUARDRAIL",
			endpoint:     "http://localhost:8003",
			capabilities: []string{"GUARDRAIL_CHECK", "RULE_VALIDATION"},
		},
		{
			name:         "Clearance Agent",
			agentType:    "CLEARANCE",
			endpoint:     "http://localhost:8004",
			capabilities: []string{"LOAN_APPROVAL", "CLEARANCE_DECISION"},
		},
		{
			name:         "Scoring Agent",
			agentType:    "SCORING",
			endpoint:     "http://localhost:8005",
			capabilities: []string{"CREDIT_SCORE", "RISK_SCORE"},
		},
	}

	for _, agentDef := range defaultAgents {
		req := &model.AgentRegistrationRequest{
			Name:         agentDef.name,
			Type:         agentDef.agentType,
			Endpoint:     agentDef.endpoint,
			Capabilities: agentDef.capabilities,
		}

		_, err := registry.RegisterAgent(ctx, req)
		if err != nil {
			log.Warn().Err(err).Str("agent", agentDef.name).Msg("Failed to register default agent")
		} else {
			log.Info().Str("agent", agentDef.name).Msg("Registered default agent")
		}
	}
}


package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aibanking/agent-mesh/internal/config"
	"github.com/aibanking/agent-mesh/internal/controller"
	"github.com/aibanking/agent-mesh/internal/middleware"
	"github.com/aibanking/agent-mesh/internal/router"
	"github.com/aibanking/agent-mesh/internal/service"
	"github.com/aibanking/agent-mesh/internal/utils"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	utils.InitLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Trim whitespace from agent type to handle trailing spaces
	agentType := strings.TrimSpace(cfg.Agent.Type)
	agentName := strings.TrimSpace(cfg.Agent.Name)
	endpoint := strings.TrimSpace(cfg.Agent.Endpoint)

	log.Info().
		Str("agent_type", agentType).
		Str("agent_name", agentName).
		Str("endpoint", endpoint).
		Msg("Starting Agent")

	// Create agent base
	agentBase := service.NewAgentBase(agentType, agentName, endpoint, &cfg.MCPServer, &cfg.MLModels, &cfg.BankingIntegrations)

	// Create specific agent based on type
	var agentProcessor service.ProcessRequest
	var capabilities []string

	switch agentType {
	case "BANKING":
		agentProcessor = service.NewBankingAgent(agentBase)
		capabilities = []string{"TRANSFER_NEFT", "TRANSFER_RTGS", "TRANSFER_IMPS", "TRANSFER_UPI", "CHECK_BALANCE", "GET_STATEMENT", "ADD_BENEFICIARY"}
	case "FRAUD":
		agentProcessor = service.NewFraudAgent(agentBase)
		capabilities = []string{"FRAUD_CHECK", "RISK_ASSESSMENT"}
	case "GUARDRAIL":
		agentProcessor = service.NewGuardrailAgent(agentBase)
		capabilities = []string{"GUARDRAIL_CHECK", "RULE_VALIDATION", "RBI_COMPLIANCE"}
	case "CLEARANCE":
		agentProcessor = service.NewClearanceAgent(agentBase)
		capabilities = []string{"LOAN_APPROVAL", "CLEARANCE_DECISION"}
	case "SCORING":
		agentProcessor = service.NewScoringAgent(agentBase)
		capabilities = []string{"CREDIT_SCORE", "FRAUD_SCORE", "RISK_SCORE"}
	default:
		log.Fatal().Str("agent_type", agentType).Msg("Unknown agent type")
	}

	// Register with MCP Server if enabled
	if cfg.Agent.AutoRegister {
		ctx := context.Background()
		if err := agentBase.RegisterWithMCP(ctx, capabilities); err != nil {
			log.Warn().Err(err).Msg("Failed to register with MCP Server, continuing anyway")
		}
	}

	// Initialize controller
	agentController := controller.NewAgentController(agentProcessor, agentType)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Initialize router
	appRouter := router.NewRouter(agentController, rateLimiter)
	r := appRouter.SetupRoutes()

	// Create HTTP server - ensure port is trimmed
	serverPort := strings.TrimSpace(cfg.Server.Port)
	serverHost := strings.TrimSpace(cfg.Server.Host)
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", serverHost, serverPort),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("address", server.Addr).
			Str("agent_type", agentType).
			Msg("Agent started")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down agent...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Agent exited")
}


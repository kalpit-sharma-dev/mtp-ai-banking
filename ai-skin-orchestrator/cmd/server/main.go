package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aibanking/ai-skin-orchestrator/internal/config"
	"github.com/aibanking/ai-skin-orchestrator/internal/controller"
	"github.com/aibanking/ai-skin-orchestrator/internal/middleware"
	"github.com/aibanking/ai-skin-orchestrator/internal/router"
	"github.com/aibanking/ai-skin-orchestrator/internal/service"
	"github.com/aibanking/ai-skin-orchestrator/internal/utils"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	utils.InitLogger(cfg.Logging.Level, cfg.Logging.Format)

	log.Info().Msg("Starting AI Skin Orchestrator (Layer 2)")

	// Initialize services
	llmService := service.NewLLMService(&cfg.LLM)
	historyService := service.NewHistoryService()
	behaviorAnalyzer := service.NewBehaviorAnalyzer()
	riskCalculator := service.NewRiskCalculator()

	intentParser := service.NewIntentParser(llmService, cfg.LLM.Enabled)
	contextEnricher := service.NewContextEnricher(historyService, behaviorAnalyzer, riskCalculator)
	mcpClient := service.NewMCPClient(&cfg.MCPServer)
	responseMerger := service.NewResponseMerger()

	orchestrator := service.NewOrchestrator(
		intentParser,
		contextEnricher,
		mcpClient,
		responseMerger,
	)

	// Initialize session service
	sessionService := service.NewSessionService()

	// Initialize controllers
	orchestratorController := controller.NewOrchestratorController(orchestrator, sessionService)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Initialize router
	appRouter := router.NewRouter(orchestratorController, rateLimiter)
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
			Str("mcp_server", cfg.MCPServer.BaseURL).
			Bool("llm_enabled", cfg.LLM.Enabled).
			Msg("AI Skin Orchestrator started")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down AI Skin Orchestrator...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("AI Skin Orchestrator exited")
}


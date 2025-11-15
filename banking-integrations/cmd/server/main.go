package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aibanking/banking-integrations/internal/config"
	"github.com/aibanking/banking-integrations/internal/controller"
	"github.com/aibanking/banking-integrations/internal/middleware"
	"github.com/aibanking/banking-integrations/internal/router"
	"github.com/aibanking/banking-integrations/internal/service"
	"github.com/aibanking/banking-integrations/internal/utils"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	utils.InitLogger(cfg.Logging.Level, cfg.Logging.Format)

	log.Info().Msg("Starting Banking Integrations Service (Layer 5)")

	// Initialize services
	mbService := service.NewMBService()
	nbService := service.NewNBService()
	dwhService := service.NewDWHService(&cfg.DWH)
	bankingGateway := service.NewBankingGateway(mbService, nbService, dwhService)

	// Initialize controller
	bankingController := controller.NewBankingController(bankingGateway)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter()

	// Initialize router
	appRouter := router.NewRouter(bankingController, rateLimiter)
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
			Bool("db_enabled", cfg.Database.Enabled).
			Bool("dwh_enabled", cfg.DWH.Enabled).
			Msg("Banking Integrations Service started")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down Banking Integrations Service...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Banking Integrations Service exited")
}


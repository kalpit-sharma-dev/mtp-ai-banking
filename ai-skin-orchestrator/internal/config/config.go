package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the AI Skin Orchestrator
type Config struct {
	Server      ServerConfig
	MCPServer   MCPServerConfig
	LLM         LLMConfig
	Context     ContextConfig
	Logging     LoggingConfig
	Security    SecurityConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// MCPServerConfig holds MCP Server connection configuration
type MCPServerConfig struct {
	BaseURL string
	APIKey  string
	Timeout int
}

// LLMConfig holds LLM service configuration
type LLMConfig struct {
	Provider    string // "openai", "anthropic", "local"
	APIKey      string
	Model       string
	BaseURL     string // For local/self-hosted models
	Temperature float64
	MaxTokens   int
	Enabled     bool
}

// ContextConfig holds context enrichment configuration
type ContextConfig struct {
	HistoryLookbackDays int
	EnableBehaviorAnalysis bool
	EnableRiskScoring   bool
	CacheTTL           int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	APIKeyHeader string
	JWTSecret    string
	RateLimitRPS int
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.SetDefault("SERVER_PORT", "8081")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("MCP_SERVER_URL", "http://localhost:8080")
	viper.SetDefault("MCP_SERVER_API_KEY", "test-api-key")
	viper.SetDefault("MCP_SERVER_TIMEOUT", "30")
	viper.SetDefault("LLM_PROVIDER", "ollama") // Default to Ollama
	viper.SetDefault("LLM_MODEL", "llama3")
	viper.SetDefault("LLM_BASE_URL", "http://localhost:11434") // Ollama default port
	viper.SetDefault("LLM_TEMPERATURE", "0.7")
	viper.SetDefault("LLM_MAX_TOKENS", "1000")
	viper.SetDefault("LLM_ENABLED", "true")
	viper.SetDefault("CONTEXT_HISTORY_DAYS", "90")
	viper.SetDefault("CONTEXT_ENABLE_BEHAVIOR", "true")
	viper.SetDefault("CONTEXT_ENABLE_RISK", "true")
	viper.SetDefault("CONTEXT_CACHE_TTL", "300")
	viper.SetDefault("LOGGING_LEVEL", "info")
	viper.SetDefault("LOGGING_FORMAT", "json")
	viper.SetDefault("SECURITY_API_KEY_HEADER", "X-API-Key")
	viper.SetDefault("SECURITY_RATE_LIMIT_RPS", "100")

	// Bind environment variables
	viper.AutomaticEnv()

	AppConfig = &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8081"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  120,
		},
		MCPServer: MCPServerConfig{
			BaseURL: getEnv("MCP_SERVER_URL", "http://localhost:8080"),
			APIKey:  getEnv("MCP_SERVER_API_KEY", "test-api-key"),
			Timeout: 30,
		},
		LLM: LLMConfig{
			Provider:    getEnv("LLM_PROVIDER", "ollama"),
			APIKey:      getEnv("LLM_API_KEY", ""),
			Model:       getEnv("LLM_MODEL", "llama3"),
			BaseURL:     getEnv("LLM_BASE_URL", "http://localhost:11434"),
			Temperature: 0.7,
			MaxTokens:   1000,
			Enabled:     getEnv("LLM_ENABLED", "true") == "true",
		},
		Context: ContextConfig{
			HistoryLookbackDays:    90,
			EnableBehaviorAnalysis: true,
			EnableRiskScoring:      true,
			CacheTTL:             300,
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOGGING_LEVEL", "info"),
			Format: getEnv("LOGGING_FORMAT", "json"),
		},
		Security: SecurityConfig{
			APIKeyHeader: getEnv("SECURITY_API_KEY_HEADER", "X-API-Key"),
			JWTSecret:    getEnv("SECURITY_JWT_SECRET", "your-secret-key"),
			RateLimitRPS: 100,
		},
	}

	return AppConfig, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


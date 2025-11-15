package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for agents
type Config struct {
	Server              ServerConfig
	MCPServer           MCPServerConfig
	Agent               AgentConfig
	Logging             LoggingConfig
	Security            SecurityConfig
	MLModels            MLModelsConfig
	BankingIntegrations BankingIntegrationsConfig
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

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	Type         string // BANKING, FRAUD, GUARDRAIL, CLEARANCE, SCORING
	Name         string
	Endpoint     string
	Capabilities []string
	AutoRegister bool
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

// MLModelsConfig holds ML Models service configuration
type MLModelsConfig struct {
	BaseURL string
	APIKey  string
	Timeout int
	Enabled bool
}

// BankingIntegrationsConfig holds Banking Integrations service configuration
type BankingIntegrationsConfig struct {
	BaseURL string
	APIKey  string
	Timeout int
	Enabled bool
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.SetDefault("SERVER_PORT", "8001")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("MCP_SERVER_URL", "http://localhost:8080")
	viper.SetDefault("MCP_SERVER_API_KEY", "test-api-key")
	viper.SetDefault("AGENT_TYPE", "BANKING")
	viper.SetDefault("AGENT_NAME", "Banking Agent")
	viper.SetDefault("AGENT_ENDPOINT", "http://localhost:8001")
	viper.SetDefault("AGENT_AUTO_REGISTER", "true")
	viper.SetDefault("LOGGING_LEVEL", "info")
	viper.SetDefault("LOGGING_FORMAT", "json")
	viper.SetDefault("SECURITY_API_KEY_HEADER", "X-API-Key")
	viper.SetDefault("SECURITY_RATE_LIMIT_RPS", "100")
	viper.SetDefault("ML_MODELS_URL", "http://localhost:9000")
	viper.SetDefault("ML_MODELS_API_KEY", "test-api-key")
	viper.SetDefault("ML_MODELS_ENABLED", "true")
	viper.SetDefault("ML_MODELS_TIMEOUT", "10")
	viper.SetDefault("BANKING_INTEGRATIONS_URL", "http://localhost:7000")
	viper.SetDefault("BANKING_INTEGRATIONS_API_KEY", "test-api-key")
	viper.SetDefault("BANKING_INTEGRATIONS_ENABLED", "true")
	viper.SetDefault("BANKING_INTEGRATIONS_TIMEOUT", "10")

	viper.AutomaticEnv()

	AppConfig = &Config{
		Server: ServerConfig{
			Port:         strings.TrimSpace(getEnv("SERVER_PORT", "8001")),
			Host:         strings.TrimSpace(getEnv("SERVER_HOST", "0.0.0.0")),
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  120,
		},
		MCPServer: MCPServerConfig{
			BaseURL: getEnv("MCP_SERVER_URL", "http://localhost:8080"),
			APIKey:  getEnv("MCP_SERVER_API_KEY", "test-api-key"),
			Timeout: 30,
		},
		Agent: AgentConfig{
			Type:         strings.TrimSpace(getEnv("AGENT_TYPE", "BANKING")),
			Name:         strings.TrimSpace(getEnv("AGENT_NAME", "Banking Agent")),
			Endpoint:     strings.TrimSpace(getEnv("AGENT_ENDPOINT", "http://localhost:8001")),
			Capabilities: []string{}, // Will be set based on agent type
			AutoRegister: getEnv("AGENT_AUTO_REGISTER", "true") == "true",
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
		MLModels: MLModelsConfig{
			BaseURL: getEnv("ML_MODELS_URL", "http://localhost:9000"),
			APIKey:  getEnv("ML_MODELS_API_KEY", "test-api-key"),
			Timeout: 10,
			Enabled: getEnv("ML_MODELS_ENABLED", "true") == "true",
		},
		BankingIntegrations: BankingIntegrationsConfig{
			BaseURL: getEnv("BANKING_INTEGRATIONS_URL", "http://localhost:7000"),
			APIKey:  getEnv("BANKING_INTEGRATIONS_API_KEY", "test-api-key"),
			Timeout: 10,
			Enabled: getEnv("BANKING_INTEGRATIONS_ENABLED", "true") == "true",
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


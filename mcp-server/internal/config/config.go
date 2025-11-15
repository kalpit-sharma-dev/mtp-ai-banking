package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Security SecurityConfig
	Logging  LoggingConfig
	Agents   AgentsConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	GRPCPort     string
	Host         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	APIKeyHeader string
	JWTSecret    string
	RateLimitRPS int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// AgentsConfig holds agent-related configuration
type AgentsConfig struct {
	DefaultTimeout int
	HealthCheckInterval int
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_GRPC_PORT", "9090")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "aibanking")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", "0")
	viper.SetDefault("SECURITY_API_KEY_HEADER", "X-API-Key")
	viper.SetDefault("SECURITY_JWT_SECRET", "your-secret-key-change-in-production")
	viper.SetDefault("SECURITY_RATE_LIMIT_RPS", "100")
	viper.SetDefault("LOGGING_LEVEL", "info")
	viper.SetDefault("LOGGING_FORMAT", "json")
	viper.SetDefault("AGENTS_DEFAULT_TIMEOUT", "30")
	viper.SetDefault("AGENTS_HEALTH_CHECK_INTERVAL", "60")

	// Bind environment variables
	viper.AutomaticEnv()

	AppConfig = &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			GRPCPort:     getEnv("SERVER_GRPC_PORT", "9090"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  120,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "aibanking"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Security: SecurityConfig{
			APIKeyHeader: getEnv("SECURITY_API_KEY_HEADER", "X-API-Key"),
			JWTSecret:    getEnv("SECURITY_JWT_SECRET", "your-secret-key-change-in-production"),
			RateLimitRPS: 100,
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOGGING_LEVEL", "info"),
			Format: getEnv("LOGGING_FORMAT", "json"),
		},
		Agents: AgentsConfig{
			DefaultTimeout:       30,
			HealthCheckInterval: 60,
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


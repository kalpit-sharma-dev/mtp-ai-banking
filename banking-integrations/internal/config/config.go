package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	DWH      DWHConfig
	Logging  LoggingConfig
	Security SecurityConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
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
	Enabled  bool
}

// DWHConfig holds Data Warehouse configuration
type DWHConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	Enabled  bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	APIKeyHeader string
	JWTSecret    string
	RateLimitRPS int
}

var AppConfig *Config

// LoadConfig loads configuration from environment
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.SetDefault("SERVER_PORT", "7000")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "banking")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_ENABLED", "false")
	viper.SetDefault("DWH_HOST", "localhost")
	viper.SetDefault("DWH_PORT", "5432")
	viper.SetDefault("DWH_USER", "postgres")
	viper.SetDefault("DWH_PASSWORD", "postgres")
	viper.SetDefault("DWH_NAME", "dwh")
	viper.SetDefault("DWH_SSLMODE", "disable")
	viper.SetDefault("DWH_ENABLED", "false")
	viper.SetDefault("LOGGING_LEVEL", "info")
	viper.SetDefault("LOGGING_FORMAT", "json")
	viper.SetDefault("SECURITY_API_KEY_HEADER", "X-API-Key")
	viper.SetDefault("SECURITY_RATE_LIMIT_RPS", "100")

	viper.AutomaticEnv()

	AppConfig = &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "7000"),
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
			DBName:   getEnv("DB_NAME", "banking"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Enabled:  getEnv("DB_ENABLED", "false") == "true",
		},
		DWH: DWHConfig{
			Host:     getEnv("DWH_HOST", "localhost"),
			Port:     getEnv("DWH_PORT", "5432"),
			User:     getEnv("DWH_USER", "postgres"),
			Password: getEnv("DWH_PASSWORD", "postgres"),
			DBName:   getEnv("DWH_NAME", "dwh"),
			SSLMode:  getEnv("DWH_SSLMODE", "disable"),
			Enabled:  getEnv("DWH_ENABLED", "false") == "true",
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


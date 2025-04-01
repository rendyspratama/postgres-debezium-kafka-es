package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Port:      getEnvOrDefault("API_PORT", "8081"),
		DBHost:    getEnvOrDefault("POSTGRES_HOST", "localhost"),
		DBPort:    getEnvOrDefault("POSTGRES_PORT", "5432"),
		DBUser:    getEnvOrDefault("POSTGRES_USER", "user"),
		DBPass:    getEnvOrDefault("POSTGRES_PASSWORD", "password"),
		DBName:    getEnvOrDefault("POSTGRES_DB", "digital_discovery"),
		DBSSLMode: getEnvOrDefault("POSTGRES_SSL_MODE", "disable"),
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPass +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

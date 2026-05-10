// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration, loaded from environment variables.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Agent    AgentConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeoutSec  int
	WriteTimeoutSec int
}

type DatabaseConfig struct {
	DSN string // PostgreSQL Data Source Name
}

type JWTConfig struct {
	Secret            string
	AccessExpiryHours int
	RefreshExpiryDays int
}

type AgentConfig struct {
	DEEPSEEKAPIKEY string
	GROKAPIKEY     string
}

// Load reads environment variables and returns a populated Config struct.
// It fails fast if required variables are missing.
func LoadConfig() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET environment variable is required")
	}

	dbDSN := os.Getenv("DATABASE_URL")
	if dbDSN == "" {
		return nil, fmt.Errorf("config: DATABASE_URL environment variable is required")
	}

	readTimeout, _ := strconv.Atoi(os.Getenv("SERVER_READ_TIMEOUT_SEC"))
	if readTimeout == 0 {
		readTimeout = 10
	}
	writeTimeout, _ := strconv.Atoi(os.Getenv("SERVER_WRITE_TIMEOUT_SEC"))
	if writeTimeout == 0 {
		writeTimeout = 30
	}
	jwtExpiry, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if jwtExpiry == 0 {
		jwtExpiry = 1
	}
	refreshTokenExpiry, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if refreshTokenExpiry == 0 {
		refreshTokenExpiry = 24
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	deepseekApiKey := os.Getenv("DEEPSEEKAPIKEY")
	grokApiKey := os.Getenv("grokApiKey")

	return &Config{
		Server: ServerConfig{
			Port:            port,
			ReadTimeoutSec:  readTimeout,
			WriteTimeoutSec: writeTimeout,
		},
		Database: DatabaseConfig{DSN: dbDSN},
		JWT:      JWTConfig{Secret: secret, AccessExpiryHours: jwtExpiry, RefreshExpiryDays: refreshTokenExpiry},
		Agent:    AgentConfig{DEEPSEEKAPIKEY: deepseekApiKey, GROKAPIKEY: grokApiKey},
	}, nil
}

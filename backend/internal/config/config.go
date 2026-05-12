package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	BackendPort      string
	AccessSecret     string
	RefreshSecret    string
	AIServiceURL     string
	AIRequestTimeout time.Duration
	StorageDir       string
	FrontendOrigin   string
	MaxUploadBytes   int64
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
}

func Load() Config {
	return Config{
		PostgresHost:     env("POSTGRES_HOST", "localhost"),
		PostgresPort:     env("POSTGRES_PORT", "5432"),
		PostgresUser:     env("POSTGRES_USER", "queryforge"),
		PostgresPassword: env("POSTGRES_PASSWORD", "queryforge"),
		PostgresDB:       env("POSTGRES_DB", "queryforge"),
		BackendPort:      env("BACKEND_PORT", "8080"),
		AccessSecret:     env("JWT_ACCESS_SECRET", "change-me-access-secret"),
		RefreshSecret:    env("JWT_REFRESH_SECRET", "change-me-refresh-secret"),
		AIServiceURL:     env("AI_SERVICE_URL", "http://localhost:8000"),
		AIRequestTimeout: time.Duration(envInt("AI_REQUEST_TIMEOUT_SECONDS", 130)) * time.Second,
		StorageDir:       env("STORAGE_DIR", "./storage"),
		FrontendOrigin:   env("FRONTEND_ORIGIN", "http://localhost:5173"),
		MaxUploadBytes:   int64(envInt("MAX_UPLOAD_MB", 50)) * 1024 * 1024,
		AccessTTL:        15 * time.Minute,
		RefreshTTL:       7 * 24 * time.Hour,
	}
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.PostgresUser, c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

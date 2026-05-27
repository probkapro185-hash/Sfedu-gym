package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env           string
	HTTPPort      string
	DatabaseURL   string
	JWTSecret     string
	JWTTokenTTL   time.Duration
	BcryptCost    int
	CORSOrigin    string
	AdminPassword string // bcrypt-хеш пароля начального администратора
}

// Load — загрузка конфигурации из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{
		Env:           getEnv("APP_ENV", "development"),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTTokenTTL:   time.Duration(getEnvInt("JWT_TOKEN_TTL_HOURS", 24)) * time.Hour,
		BcryptCost:    getEnvInt("BCRYPT_COST", 12),
		CORSOrigin:    getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:3000"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.AdminPassword == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD is required")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

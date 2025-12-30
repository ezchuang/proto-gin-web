package config

import (
	"os"
	"strconv"
)

type Config struct {
	Env     string
	Port    string
	LogFile string

	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string

	BaseURL         string
	SiteName        string
	SiteDescription string

	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	SessionCookieName  string
	RememberCookieName string
}

func Load() Config {
	redisDB := getEnvInt("REDIS_DB", 0)
	return Config{
		Env:                getEnv("APP_ENV", "development"),
		Port:               getEnv("PORT", "8080"),
		LogFile:            getEnv("LOG_FILE", "logs/app.log"),
		DBUser:             getEnv("POSTGRES_USER", "proto_user"),
		DBPassword:         getEnv("POSTGRES_PASSWORD", "proto_password"),
		DBName:             getEnv("POSTGRES_DB", "proto_db"),
		DBHost:             getEnv("POSTGRES_HOST", "localhost"),
		DBPort:             getEnv("POSTGRES_PORT", "5432"),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
		SiteName:           getEnv("SITE_NAME", "Prototype"),
		SiteDescription:    getEnv("SITE_DESCRIPTION", "Gin + SSR + SQLC prototype"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		SessionCookieName:  getEnv("ADMIN_SESSION_COOKIE", "admin_session"),
		RememberCookieName: getEnv("ADMIN_REMEMBER_COOKIE", "admin_remember"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

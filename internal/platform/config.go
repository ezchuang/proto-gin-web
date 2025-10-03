package platform

import "os"

type Config struct {
	Env  string
	Port string

	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string

	BaseURL         string
	SiteName        string
	SiteDescription string

	// Admin basic credentials (stub)
	AdminUser string
	AdminPass string
}

func Load() Config {
	return Config{
		Env:             getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DBUser:          getEnv("POSTGRES_USER", "proto_user"),
		DBPassword:      getEnv("POSTGRES_PASSWORD", "proto_password"),
		DBName:          getEnv("POSTGRES_DB", "proto_db"),
		DBHost:          getEnv("POSTGRES_HOST", "localhost"),
		DBPort:          getEnv("POSTGRES_PORT", "5432"),
		BaseURL:         getEnv("BASE_URL", "http://localhost:8080"),
		SiteName:        getEnv("SITE_NAME", "Proto"),
		SiteDescription: getEnv("SITE_DESCRIPTION", "Gin + SSR + SQLC prototype"),
		AdminUser:       getEnv("ADMIN_USER", "admin"),
		AdminPass:       getEnv("ADMIN_PASS", "changeme"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

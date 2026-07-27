package platform

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DatabaseURL        string
	Env                string
	HTTPAddr           string
	GRPCAddr           string
	GRPCTLSCertFile    string
	GRPCTLSKeyFile     string
	JWTSecret          string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
	FrontendURL        string
}

func LoadConfig() Config {
	return Config{
		DatabaseURL:        requireEnv("DATABASE_URL"),
		Env:                getEnv("ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		GRPCAddr:           getEnv("GRPC_ADDR", ":50051"),
		GRPCTLSCertFile:    getEnv("GRPC_TLS_CERT_FILE", ""),
		GRPCTLSKeyFile:     getEnv("GRPC_TLS_KEY_FILE", ""),
		JWTSecret:          requireEnv("JWT_SECRET"),
		GitHubClientID:     requireEnv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: requireEnv("GITHUB_CLIENT_SECRET"),
		GitHubCallbackURL:  getEnv("GITHUB_CALLBACK_URL", "http://localhost:8080/auth/github/callback"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:4200"),
	}
}

func NewDB(cfg Config) (*gorm.DB, error) {
	l := logger.Default.LogMode(logger.Info)
	if cfg.Env == "production" {
		l = logger.Default.LogMode(logger.Error)
	}
	return gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{Logger: l})
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %q is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

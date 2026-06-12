package platform

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DatabaseURL string
	Env         string
	HTTPAddr    string
	GRPCAddr    string
}

func LoadConfig() Config {
	return Config{
		DatabaseURL: requireEnv("DATABASE_URL"),
		Env:         getEnv("ENV", "development"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		GRPCAddr:    getEnv("GRPC_ADDR", ":50051"),
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

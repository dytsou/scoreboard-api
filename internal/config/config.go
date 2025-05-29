package config

import (
	"os"
)

type Config struct {
	Host                     string `yaml:"host"                       envconfig:"HOST"`
	Port                     string `yaml:"port"                       envconfig:"PORT"`
	DatabaseURL              string `yaml:"database_url"               envconfig:"DATABASE_URL"`
	MigrationSource          string `yaml:"migration_source"           envconfig:"MIGRATION_SOURCE"`
	BaseURL                  string `yaml:"base_url"                   envconfig:"BASE_URL"`
	GoogleOauthClientID      string `yaml:"google_oauth_client_id"     envconfig:"GOOGLE_OAUTH_CLIENT_ID"`
	GoogleOauthClientSecret  string `yaml:"google_oauth_client_secret" envconfig:"GOOGLE_OAUTH_CLIENT_SECRET"`
}

func Load() Config {
	config := &Config{
		Host:                     getEnv("HOST", "localhost"),
		Port:                     getEnv("PORT", "8080"),
		DatabaseURL:              getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"),
		MigrationSource:          getEnv("MIGRATION_SOURCE", "file://internal/database/migrations"),
		BaseURL:                  getEnv("BASE_URL", "http://localhost:8080"),
		GoogleOauthClientID:      getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleOauthClientSecret:  getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
	}
	return *config
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

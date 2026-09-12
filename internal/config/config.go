package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string            `validate:"required"`
	Server      ServerConfig      `validate:"required"`
	Database    DatabaseConfig    `validate:"required"`
	Redis       RedisConfig       `validate:"required"`
	Integration IntegrationConfig `validate:"required"`
	OAuth       OAuthConfig       `validate:"required"`
	GCS         GCSConfig         `validate:"required"`
}

type ServerConfig struct {
	Port string `validate:"required"`
}

type DatabaseConfig struct {
	Host     string `validate:"required"`
	Port     string `validate:"required,numeric"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	Name     string `validate:"required"`
	SSLMode  string `validate:"required"`
}

type RedisConfig struct {
	// Address  string `validate:"required"`
	// Password string `validate:"required"`
	LockTTL  string `validate:"required"`
	RedisURL string `validate:"required"`
}

type IntegrationConfig struct {
	SMTPHost     string `validate:"omitempty"`
	SMTPPort     string `validate:"omitempty"`
	ResendAPIKey string `validate:"omitempty"`
}

type OAuthConfig struct {
	GoogleClientID     string `validate:"required"`
	GoogleClientSecret string `validate:"required"`
	GoogleRedirectURL  string `validate:"required"`
}

type GCSConfig struct {
	GoogleCredentialPath string `validate:"omitempty"`
	GCSBucketName        string `validate:"required"`
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func validateConfig(cnf *Config) error {
	validate := validator.New()
	err := validate.Struct(cnf)

	if err != nil {
		return err
	}
	return nil
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using system environment variables")
		// os.Exit(1)
	}

	conf := &Config{
		Env: getEnv("ENV", "development"),
		Server: ServerConfig{
			Port: getEnv("PORT", "8000"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			// Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			// Password: getEnv("REDIS_PASSWORD", ""),
			LockTTL:  getEnv("LOCK_TTL", "60000"),
			RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		Integration: IntegrationConfig{
			ResendAPIKey: getEnv("INTEGRATION_RESEND_API_KEY", ""),
			SMTPHost:     getEnv("SMTP_HOST", "localhost"),
			SMTPPort:     getEnv("SMTP_PORT", "1025"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
		},
		GCS: GCSConfig{
			GoogleCredentialPath: getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
			GCSBucketName:        getEnv("GCS_BUCKET_NAME", ""),
		},
	}

	if err := validateConfig(conf); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return conf, nil
}

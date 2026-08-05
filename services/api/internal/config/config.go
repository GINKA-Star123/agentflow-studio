package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv   string
	HTTPPort string

	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	AIRuntimeURL  string

	JWTSecret         string
	JWTIssuer         string
	JWTAccessTokenTTL time.Duration

	OTELExporterOTLPEndpoint string
}

func Load() (*Config, error) {
	_ = godotenv.Load("../../.env", ".env")

	viper.SetDefault("APP_ENV", "dev")
	viper.SetDefault("API_HTTP_PORT", "8080")

	viper.SetDefault("DATABASE_URL", "postgres://agentflow:agentflow@localhost:5432/agentflow?sslmode=disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("AI_RUNTIME_URL", "http://localhost:8090")

	viper.SetDefault("JWT_SECRET", "change-me")
	viper.SetDefault("JWT_ISSUER", "agentflow-studio")
	viper.SetDefault("JWT_ACCESS_TOKEN_TTL", "2h")

	viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")

	viper.AutomaticEnv()

	accessTokenTTL, err := time.ParseDuration(viper.GetString("JWT_ACCESS_TOKEN_TTL"))
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TOKEN_TTL 格式错误: %w", err)
	}

	cfg := &Config{
		AppEnv:      viper.GetString("APP_ENV"),
		HTTPPort:    viper.GetString("API_HTTP_PORT"),
		DatabaseURL: viper.GetString("DATABASE_URL"),

		RedisAddr:     viper.GetString("REDIS_ADDR"),
		RedisPassword: viper.GetString("REDIS_PASSWORD"),
		AIRuntimeURL:  viper.GetString("AI_RUNTIME_URL"),

		JWTSecret:         viper.GetString("JWT_SECRET"),
		JWTIssuer:         viper.GetString("JWT_ISSUER"),
		JWTAccessTokenTTL: accessTokenTTL,

		OTELExporterOTLPEndpoint: viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}

	if cfg.HTTPPort == "" {
		return nil, fmt.Errorf("API_HTTP_PORT 不能为空")
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL 不能为空")
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "change-me" {
		if cfg.AppEnv == "prod" {
			return nil, fmt.Errorf("生产环境必须配置安全的 JWT_SECRET")
		}
	}

	if cfg.JWTIssuer == "" {
		return nil, fmt.Errorf("JWT_ISSUER 不能为空")
	}

	if cfg.JWTAccessTokenTTL <= 0 {
		return nil, fmt.Errorf("JWT_ACCESS_TOKEN_TTL 必须大于 0")
	}

	return cfg, nil
}

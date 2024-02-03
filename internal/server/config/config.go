package config

import (
	"time"
)

type Config struct {
	LogLevel string `toml:"log_level" env-required:"true"`
	Port     string `toml:"port" env-required:"true"`

	// APIKey for accessing from telegram bot
	APIKey string `env:"SERVER_API_KEY" env-required:"true"`

	PostgresURL string      `env:"PG_URL" env-required:"true"`
	AccessToken AccessToken `toml:"access_token"`
}

type AccessToken struct {
	SecretKey string        `env:"SERVER_ACCCESS_TOKEN_SECRET" env-required:"true"`
	TTL       time.Duration `toml:"ttl" env-required:"true"`
}

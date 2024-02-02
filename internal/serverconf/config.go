package serverconf

import (
	"log/slog"
	"strings"
)

type Config struct {
	LogLevel string `toml:"log_level" env-required:"true"`
	Port     string `toml:"port" env-required:"true"`

	// APIKey for accessing from telegram bot
	APIKey string `env:"SERVER_API_KEY" env-required:"true"`

	PostgresURL string `env:"PG_URL" env-required:"true"`
}

func (c *Config) SlogLogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

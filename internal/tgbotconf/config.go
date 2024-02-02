package tgbotconf

import (
	"log/slog"
	"strings"
)

type Config struct {
	LogLevel     string `toml:"log_level" env-required:"true"`
	AccessToken  string `env:"TGBOT_ACCESS_TOKEN" env-required:"true"`
	ServerURL    string `toml:"server_url" env-required:"true"`
	ServerAPIKey string `env:"SERVER_API_KEY" env-required:"true"`
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

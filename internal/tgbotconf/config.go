package tgbotconf

import (
	"log/slog"
	"strings"
	"time"
)

type Config struct {
	LogLevel    string `toml:"log_level" env-required:"true"`
	AccessToken string `env:"TGBOT_ACCESS_TOKEN" env-required:"true"`
	Server      Server `toml:"server"`
	Redis       Redis  `toml:"redis"`
}

type Server struct {
	URL           string        `toml:"url" env-required:"true"`
	APIKey        string        `env:"SERVER_API_KEY" env-required:"true"`
	ClientTimeout time.Duration `toml:"client_timeout" env-required:"true"`
}

type Redis struct {
	Addr         string        `toml:"addr" env-required:"true"`
	Password     string        `env:"REDIS_PASSWORD" env-required:"true"`
	ReadTimeout  time.Duration `toml:"read_timeout" env-required:"true"`
	WriteTimeout time.Duration `toml:"write_timeout" env-required:"true"`
	PoolSize     int           `toml:"pool_size" env-required:"true"`
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

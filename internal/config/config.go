package config

type Config struct {
	LogLevel    string   `toml:"log_level" env-required:"true"`
	AccessToken string   `env:"TELEGRAM_ACCESS_TOKEN" env-required:"true"`
	Verbose     bool     `toml:"verbose"`
	Postgres    Postgres `toml:"postgres"`
	CommitHash  string
}

type Postgres struct {
	URL      string `env:"PG_URL" env-required:"true"`
	MaxConns int32  `toml:"max_conns" env-required:"true"`
}

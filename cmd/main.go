package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/jackc/pgx/v5/stdlib" // for goose running migrations via pgx
	"github.com/pressly/goose/v3"

	"github.com/ysomad/financer/internal/bot"
	"github.com/ysomad/financer/internal/bot/state"
	"github.com/ysomad/financer/internal/config"
	"github.com/ysomad/financer/internal/postgres"
	"github.com/ysomad/financer/internal/postgres/pgclient"
	"github.com/ysomad/financer/internal/service"
	"github.com/ysomad/financer/internal/slogx"
)

func mustMigrate(dsn, migrationsDir string) {
	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		slogx.Fatal(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			slogx.Fatal(err.Error())
		}
	}()

	if err := goose.RunContext(context.Background(), "up", db, migrationsDir); err != nil {
		slogx.Fatal(err.Error())
	}
}

func main() {
	var (
		migrate       bool
		migrationsDir string
		configPath    string
	)

	flag.BoolVar(&migrate, "migrate", false, "run migrations on start")
	flag.StringVar(&migrationsDir, "migrations-dir", "./migrations", "path to migrations directory")
	flag.StringVar(&configPath, "conf", "./configs/local.toml", "path to app config")
	flag.Parse()

	var conf config.Config

	if err := cleanenv.ReadConfig("configs/local.toml", &conf); err != nil {
		slogx.Fatal("config parse error", "err", err.Error())
	}

	if migrate {
		mustMigrate(conf.Postgres.URL, migrationsDir)
	}

	// logger
	handler := slog.Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		// AddSource: true,
		Level: slogx.ParseLevel(conf.LogLevel),
	}))
	handler = slogx.NewHandlerMiddleware(handler)
	slog.SetDefault(slog.New(handler))

	slog.Debug("loaded config", "config", conf)

	// postgres
	pgClient, err := pgclient.New(conf.Postgres.URL, pgclient.WithMaxConns(conf.Postgres.MaxConns))
	if err != nil {
		slogx.Fatal(err.Error())
	}

	categoryStorage := postgres.CategoryStorage{Client: pgClient}
	userStorage := postgres.UserStorage{Client: pgClient}
	operationStorage := postgres.OperationStorage{Client: pgClient}
	keywordStorage := postgres.KeywordStorage{Client: pgClient}

	stateStorage := expirable.NewLRU[string, state.State](100, nil, time.Hour*24)

	userService := service.NewUser(userStorage)

	bot, err := bot.New(conf, stateStorage, categoryStorage, userService, operationStorage, keywordStorage)
	if err != nil {
		slogx.Fatal(err.Error())
	}
	defer bot.Stop()

	bot.Start()
}

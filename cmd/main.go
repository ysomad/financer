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
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"github.com/ysomad/financer/internal/bot"
	"github.com/ysomad/financer/internal/config"
	"github.com/ysomad/financer/internal/domain"
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
		commitHash    string
	)

	flag.BoolVar(&migrate, "migrate", false, "run migrations on start")
	flag.StringVar(&migrationsDir, "migrations-dir", "./migrations", "path to migrations directory")
	flag.StringVar(&configPath, "conf", "./configs/local.toml", "path to app config")
	flag.StringVar(&commitHash, "commit", "0", "commit hash from git")
	flag.Parse()

	var conf config.Config

	if err := cleanenv.ReadConfig("configs/local.toml", &conf); err != nil {
		slogx.Fatal("config parse error", "err", err.Error())
	}

	conf.CommitHash = commitHash

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

	stateStorage := expirable.NewLRU[string, domain.State](100, nil, time.Hour*24)

	userService := service.NewUser(userStorage)

	b := bot.New(stateStorage, categoryStorage, userService, operationStorage, keywordStorage)

	// telegram
	tbot, err := tele.NewBot(tele.Settings{
		Token: conf.AccessToken,

		// TODO: change to webhook
		Poller:    &tele.LongPoller{Timeout: time.Second},
		Verbose:   conf.Verbose,
		OnError:   b.HandleError,
		ParseMode: tele.ModeHTML,
	})
	if err != nil {
		slogx.Fatal(err.Error())
	}
	defer tbot.Close()

	if err := tbot.SetCommands([]tele.Command{
		{
			Text:        "categories",
			Description: "List categories",
		},

		{
			Text:        "add_category",
			Description: "Add new category",
		},
		{
			Text:        "rename_category",
			Description: "Rename existing category",
		},
		{
			Text:        "set_language",
			Description: "Set bot language",
		},
		{
			Text:        "set_currency",
			Description: "Set default currency",
		},
	}); err != nil {
		slogx.Fatal(err.Error())
	}

	tbot.Use(middleware.Recover())
	tbot.Use(bot.ContextMiddleware(conf.CommitHash))
	tbot.Use(b.UserContextMiddleware)

	tbot.Handle("/start", b.Start)

	tbot.Handle("/categories", b.ListCategories)
	tbot.Handle("/rename_category", b.RenameCategory)
	tbot.Handle("/add_category", b.AddCategory)

	tbot.Handle("/set_language", b.SetLanguage)
	tbot.Handle("/set_currency", b.SetCurrency)

	tbot.Handle(tele.OnCallback, b.HandleCallback)
	tbot.Handle(tele.OnText, b.HandleText)

	tbot.Start()
}

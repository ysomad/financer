package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/jackc/pgx/v5/stdlib" // for goose running migrations via pgx
	"github.com/pressly/goose/v3"

	"github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/httpserver"
	"github.com/ysomad/financer/internal/postgres"
	"github.com/ysomad/financer/internal/postgres/pgclient"
	"github.com/ysomad/financer/internal/rpc"
	v1 "github.com/ysomad/financer/internal/rpc/telegram/v1"
	"github.com/ysomad/financer/internal/serverconf"
	"github.com/ysomad/financer/internal/slogx"
)

func main() {
	var (
		migrate       bool
		migrationsDir string
		configPath    string
	)

	flag.BoolVar(&migrate, "migrate", false, "run migrations on start")
	flag.StringVar(&migrationsDir, "migrations-dir", "./migrations", "path to migrations directory")
	flag.StringVar(&configPath, "conf", "./configs/server_local.toml", "path to app config")
	flag.Parse()

	var conf serverconf.Config

	if err := cleanenv.ReadConfig(configPath, &conf); err != nil {
		log.Fatalf("config parse error: %s", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: conf.SlogLogLevel(),
	})))

	slog.Debug("loaded config", "conf", conf)

	if migrate {
		mustMigrate(conf.PostgresURL, migrationsDir)
	}

	// postgres
	pgclient, err := pgclient.New(conf.PostgresURL)
	if err != nil {
		slogx.Fatal("postgres client not created", err)
	}

	identityStorage := postgres.NewIdentityStorage(pgclient)

	// connect
	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		slogx.Fatal("validate interceptor not created", err)
	}

	tgInterceptor := rpc.NewAPIKeyInterceptor(conf.APIKey)

	mux := http.NewServeMux()

	// identity service
	identitysrv := v1.NewIdentityServer(identityStorage)
	path, handler := telegramv1connect.NewIdentityServiceHandler(identitysrv,
		connect.WithInterceptors(validateInterceptor, tgInterceptor))
	mux.Handle(path, handler)

	// access token service
	tokensrv := v1.NewTokenServer(identityStorage, conf.AccessToken)
	path, handler = telegramv1connect.NewAccessTokenServiceHandler(tokensrv,
		connect.WithInterceptors(validateInterceptor, tgInterceptor))
	mux.Handle(path, handler)

	srv := httpserver.New(mux, httpserver.WithAddr("0.0.0.0", conf.Port))
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	select {
	case s := <-interrupt:
		slog.Info("received interrupt signal", "signal", s.String())
	case err := <-srv.Notify():
		slog.Error("got error from http server", "error", err.Error())
	}

	if err := srv.Shutdown(); err != nil {
		slog.Error("got error on http server shutdown", "error", err.Error())
	}
}

func mustMigrate(dsn, migrationsDir string) {
	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	if err := goose.RunContext(context.Background(), "up", db, migrationsDir); err != nil {
		panic(err)
	}
}

package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/ilyakaznacheev/cleanenv"
	goredis "github.com/redis/go-redis/v9"
	tele "gopkg.in/telebot.v3"

	"github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/slogx"
	"github.com/ysomad/financer/internal/tgbot"
	"github.com/ysomad/financer/internal/tgbot/config"
	"github.com/ysomad/financer/internal/tgbot/redis"
)

func main() {
	var conf config.Config

	if err := cleanenv.ReadConfig("configs/tgbot_local.toml", &conf); err != nil {
		log.Fatalf("config parse error: %s", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogx.ParseLevel(conf.LogLevel),
	})))

	slog.Debug("loaded config", "conf", conf)

	// redis
	rdb := goredis.NewClient(&goredis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		ReadTimeout:  conf.Redis.ReadTimeout,
		WriteTimeout: conf.Redis.WriteTimeout,
		PoolSize:     conf.Redis.PoolSize,
	})

	cmd := rdb.Ping(context.Background())
	err := cmd.Err()
	if err != nil {
		slogx.Fatal("redis client not created", err)
	}

	identityCache := redis.NewIdentityCache(rdb, conf.Cache.IdentityTTL)
	stateCache := redis.NewStateCache(rdb, conf.Cache.StateTTL)

	// telegram
	b, err := tele.NewBot(tele.Settings{
		Token: conf.AccessToken,

		// TODO: change to webhook
		Poller:  &tele.LongPoller{Timeout: time.Second},
		Verbose: conf.Debug,
	})
	if err != nil {
		slogx.Fatal("telebot not created", err)
	}

	httpClient := &http.Client{
		Timeout: conf.Server.ClientTimeout,
	}

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		slogx.Fatal("validate interceptor not created", err)
	}

	identityClient := telegramv1connect.NewIdentityServiceClient(
		httpClient,
		conf.Server.URL,
		connect.WithHTTPGet(),
		connect.WithInterceptors(validateInterceptor))

	accessTokenClient := telegramv1connect.NewAccessTokenServiceClient(
		httpClient,
		conf.Server.URL,
		connect.WithInterceptors(validateInterceptor))

	bot := tgbot.New(conf, rdb, tgbot.IdentityService{
		Client: identityClient,
		Cache:  identityCache,
	}, accessTokenClient, stateCache)

	b.Handle("/start", bot.Start)
	b.Handle("/set_currency", bot.CmdSetCurrency)

	b.Handle(tele.OnCallback, bot.HandleCallback)
	b.Start()
}

func newServerRequest[T any](msg *T, apiKey string) *connect.Request[T] {
	r := connect.NewRequest(msg)
	r.Header().Set("X-API-KEY", apiKey)
	return r
}

package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/ilyakaznacheev/cleanenv"
	tele "gopkg.in/telebot.v3"

	"github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/slogx"
	"github.com/ysomad/financer/internal/tgbot"
	"github.com/ysomad/financer/internal/tgbotconf"
)

func main() {
	var conf tgbotconf.Config

	if err := cleanenv.ReadConfig("configs/tgbot_local.toml", &conf); err != nil {
		log.Fatalf("config parse error: %s", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: conf.SlogLogLevel(),
	})))

	slog.Debug("loaded config", "conf", conf)

	// TODO: change to webhook
	b, err := tele.NewBot(tele.Settings{
		Token:  conf.AccessToken,
		Poller: &tele.LongPoller{Timeout: time.Second},
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

	identityClient := telegramv1connect.NewIdentityServiceClient(httpClient,
		conf.Server.URL, connect.WithHTTPGet(), connect.WithInterceptors(validateInterceptor))

	accessTokenClient := telegramv1connect.NewAccessTokenServiceClient(httpClient, conf.Server.URL,
		connect.WithInterceptors(validateInterceptor))

	bot := tgbot.NewBot(conf, identityClient, accessTokenClient)

	b.Handle("/start", bot.HandleStart)

	b.Start()
}

func newServerRequest[T any](msg *T, apiKey string) *connect.Request[T] {
	r := connect.NewRequest(msg)
	r.Header().Set("X-API-KEY", apiKey)
	return r
}

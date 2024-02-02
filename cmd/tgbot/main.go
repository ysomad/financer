package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/ilyakaznacheev/cleanenv"
	tele "gopkg.in/telebot.v3"

	pb "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	tgv1 "github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
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

	b, err := tele.NewBot(tele.Settings{
		Token:  conf.AccessToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	identity := tgv1.NewIdentityServiceClient(http.DefaultClient,
		conf.ServerURL, connect.WithHTTPGet())

	b.Handle("/start", func(c tele.Context) error {
		telegramUID := c.Chat().ID
		ctx := context.Background()

		req := newServerRequest(&pb.GetIdentityByTelegramUIDRequest{TgUid: telegramUID}, conf.ServerAPIKey)

		res, err := identity.GetIdentityByTelegramUID(ctx, req)
		// user with telegram id found
		if err == nil {
			slog.Info("identity found", "tg_uid", res.Msg.TgUid, "id", res.Msg.Id)
			return c.Send(res.Msg.GetId())
		}

		// user with telegram id not found
		if connectErr := new(connect.Error); errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
			req := newServerRequest(&pb.CreateIdentityRequest{TgUid: telegramUID}, conf.ServerAPIKey)

			res, err := identity.CreateIdentity(ctx, req)
			if err != nil {
				slog.Error("identity not created", "err", err.Error(), "uid", telegramUID)
				return c.Send("Я поднаебнулся, пробуй позже")
			}

			slog.Info("identity created", "tg_uid", res.Msg.TgUid, "id", res.Msg.Id)
		}

		return nil
	})

	b.Start()
}

func newServerRequest[T any](msg *T, apiKey string) *connect.Request[T] {
	r := connect.NewRequest(msg)
	r.Header().Set("X-API-KEY", apiKey)
	return r
}

package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"gopkg.in/telebot.v3"

	pb "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	connectpb "github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/tgbotconf"
)

type Bot struct {
	conf     tgbotconf.Config
	identity connectpb.IdentityServiceClient
	token    connectpb.AccessTokenServiceClient
}

const msgInternal = "Чет я поднаебнулся, попробуй позже..."

func NewBot(conf tgbotconf.Config, id connectpb.IdentityServiceClient, t connectpb.AccessTokenServiceClient) *Bot {
	return &Bot{conf: conf, identity: id, token: t}
}

func (b *Bot) HandleStart(c telebot.Context) error {
	tgUID := c.Chat().ID
	ctx := context.Background()

	identity, err := b.getOrCreateIdentity(ctx, tgUID)
	if err != nil {
		slog.Error("identity not found or not created", "err", err.Error())
		return c.Send(msgInternal)
	}

	c.Send(identity)

	req := newServerRequest(&pb.IssueAccessTokenRequest{TgUid: tgUID}, b.conf.Server.APIKey)

	resp, err := b.token.IssueAccessToken(ctx, req)
	if err != nil {
		slog.Error("access token not issued", "err", err.Error())
		return c.Send(msgInternal)
	}

	return c.Send(resp.Msg.AccessToken)
}

func (b *Bot) getOrCreateIdentity(ctx context.Context, tgUID int64) (*pb.Identity, error) {
	req := newServerRequest(&pb.GetIdentityByTelegramUIDRequest{TgUid: tgUID}, b.conf.Server.APIKey)

	// found
	resp, err := b.identity.GetIdentityByTelegramUID(ctx, req)
	if err == nil {
		return resp.Msg, nil
	}

	// not found
	if connectErr := new(connect.Error); errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
		req := newServerRequest(&pb.CreateIdentityRequest{TgUid: tgUID}, b.conf.Server.APIKey)

		res, err := b.identity.CreateIdentity(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("identity not created: %w", err)
		}

		return res.Msg, nil
	}

	return nil, errors.New("unsupported error type")
}

func newServerRequest[T any](msg *T, apiKey string) *connect.Request[T] {
	r := connect.NewRequest(msg)
	r.Header().Set("X-API-KEY", apiKey)
	return r
}

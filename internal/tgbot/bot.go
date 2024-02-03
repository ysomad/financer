package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"gopkg.in/telebot.v3"

	"github.com/redis/go-redis/v9"
	pb "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	connectpb "github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/tgbot/config"
)

type Bot struct {
	conf     config.Config
	redis    *redis.Client
	identity connectpb.IdentityServiceClient
	token    connectpb.AccessTokenServiceClient
}

const msgInternal = "Чет я поднаебнулся, попробуй позже..."

var cacheTTL = time.Minute * 5

func New(conf config.Config, r *redis.Client, id connectpb.IdentityServiceClient, t connectpb.AccessTokenServiceClient) *Bot {
	return &Bot{conf: conf, identity: id, token: t, redis: r}
}

func accessTokenKey(tgUID int64) string {
	return fmt.Sprintf("access_token:%d", tgUID)
}

func (b *Bot) HandleStart(c telebot.Context) error {
	tgUID := c.Chat().ID
	ctx := context.Background()

	// get access token from cache
	accessToken := b.redis.Get(ctx, accessTokenKey(tgUID)).Val()
	if accessToken != "" {
		return c.Send(accessToken)
	}

	// issue access token and set to cache
	if _, err := b.getOrCreateIdentity(ctx, tgUID); err != nil {
		slog.Error("identity not found nor created", "err", err.Error())
		return c.Send(msgInternal)
	}

	req := newServerRequest(&pb.IssueAccessTokenRequest{TgUid: tgUID}, b.conf.Server.APIKey)

	resp, err := b.token.IssueAccessToken(ctx, req)
	if err != nil {
		slog.Error("access token not issued", "err", err.Error())
		return c.Send(msgInternal)
	}

	if err := b.redis.Set(ctx, accessTokenKey(tgUID), resp.Msg.AccessToken, cacheTTL).Err(); err != nil {
		slog.Error("access key not saved to cache")
		return nil
	}

	/* send user instruction with menu
	   - /add {amount money} {?iso 4217 currency} {comment} {date in format 20.05 or 20.05.1999} - adds expense
	*/

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

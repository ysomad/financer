package v1

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"

	pb "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	"github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/postgres"
	"github.com/ysomad/financer/internal/serverconf"
)

var _ telegramv1connect.AccessTokenServiceHandler = &TokenServer{}

const AudTelegram = "TG"

type TokenServer struct {
	identity *postgres.IdentityStorage
	conf     serverconf.AccessToken
}

func NewTokenServer(id *postgres.IdentityStorage, conf serverconf.AccessToken) *TokenServer {
	return &TokenServer{identity: id, conf: conf}
}

func (s *TokenServer) IssueAccessToken(ctx context.Context,
	r *connect.Request[pb.IssueAccessTokenRequest],
) (*connect.Response[pb.IssueAccessTokenResponse], error) {
	identity, err := s.identity.FindByTelegramUID(ctx, r.Msg.TgUid)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"aud":    AudTelegram,
		"sub":    identity.ID,
		"tg_sub": identity.TelegramUID,
		"exp":    time.Now().Add(s.conf.TTL).Unix(),
	})

	tokenstr, err := token.SignedString([]byte(s.conf.SecretKey))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(&pb.IssueAccessTokenResponse{
		AccessToken: tokenstr,
	})

	resp.Header().Set("Authorization", fmt.Sprintf("Bearer %s", tokenstr))

	return resp, nil
}

package v1

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"

	pb "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	"github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/postgres"
	"github.com/ysomad/financer/internal/server/config"
)

var _ telegramv1connect.AccessTokenServiceHandler = &AccessTokenServer{}

const AudTelegram = "TG"

type AccessTokenServer struct {
	identity postgres.IdentityStorage
	conf     config.AccessToken
}

func NewAccessTokenServer(id postgres.IdentityStorage, conf config.AccessToken) *AccessTokenServer {
	return &AccessTokenServer{identity: id, conf: conf}
}

func (s *AccessTokenServer) IssueAccessToken(ctx context.Context,
	r *connect.Request[pb.IssueAccessTokenRequest],
) (*connect.Response[pb.IssueAccessTokenResponse], error) {
	identity, err := s.identity.FindByTelegramUID(ctx, r.Msg.TgUid)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"aud":    AudTelegram,
		"sub":    identity.ID,
		"tg_sub": identity.TelegramUID.Int64,
		"exp":    time.Now().Add(s.conf.TTL).Unix(),
	})

	tokenstr, err := token.SignedString([]byte(s.conf.SecretKey))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.IssueAccessTokenResponse{
		AccessToken: tokenstr,
	}), nil
}

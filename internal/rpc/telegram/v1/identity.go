package v1

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"

	pb "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	connectpb "github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/guid"
	"github.com/ysomad/financer/internal/postgres"
)

var _ connectpb.IdentityServiceHandler = &IdentityServer{}

type IdentityServer struct {
	identity *postgres.IdentityStorage
}

func NewIdentityServer(id *postgres.IdentityStorage) *IdentityServer {
	return &IdentityServer{identity: id}
}

const defaultCurrency = "RUB"

func (s *IdentityServer) CreateIdentity(ctx context.Context,
	r *connect.Request[pb.CreateIdentityRequest],
) (*connect.Response[pb.Identity], error) {
	identityID := guid.New("identity")

	err := s.identity.Insert(ctx, postgres.InsertIdentityIn{
		ID:          identityID,
		CreatedAt:   time.Now(),
		TelegramUID: r.Msg.TgUid,
		Currency:    defaultCurrency,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.Identity{
		Id:       identityID,
		TgUid:    r.Msg.TgUid,
		Currency: defaultCurrency,
	}), nil
}

func (s *IdentityServer) GetIdentityByTelegramUID(ctx context.Context,
	r *connect.Request[pb.GetIdentityByTelegramUIDRequest],
) (*connect.Response[pb.Identity], error) {
	identity, err := s.identity.FindByTelegramUID(ctx, r.Msg.TgUid)
	if err != nil {
		if errors.Is(err, postgres.ErrIdentityNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, postgres.ErrIdentityNotFound)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.Identity{
		Id:       identity.ID,
		TgUid:    identity.TelegramUID.Int64,
		Currency: identity.Currency,
	}), nil
}

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
	identity postgres.IdentityStorage
}

func NewIdentityServer(id postgres.IdentityStorage) *IdentityServer {
	return &IdentityServer{identity: id}
}

const (
	defaultCurrency = "RUB"
)

func (s *IdentityServer) CreateIdentity(ctx context.Context, r *connect.Request[pb.CreateIdentityRequest]) (*connect.Response[pb.Identity], error) {
	identityID := guid.New("identity")

	err := s.identity.Insert(ctx, postgres.InsertIdentityParams{
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

func (s *IdentityServer) GetIdentity(ctx context.Context, r *connect.Request[pb.GetIdentityRequest]) (*connect.Response[pb.Identity], error) {
	identity, err := s.identity.FindByTelegramUID(ctx, r.Msg.TgUid)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, postgres.ErrNotFound)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.Identity{
		Id:       identity.ID,
		TgUid:    identity.TelegramUID.Int64,
		Currency: identity.Currency,
	}), nil
}

func (s *IdentityServer) UpdateIdentity(ctx context.Context, r *connect.Request[pb.UpdateIdentityRequest]) (*connect.Response[pb.Identity], error) {
	if err := s.identity.Update(ctx, postgres.UpdateIdentityParams{
		IdentityID: r.Msg.Id,
		Currency:   r.Msg.Currency,
		UpdatedAt:  time.Now(),
	}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	identity, err := s.identity.Get(ctx, r.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.Identity{
		Id:       identity.ID,
		TgUid:    identity.TelegramUID.Int64,
		Currency: identity.Currency,
	}), nil
}

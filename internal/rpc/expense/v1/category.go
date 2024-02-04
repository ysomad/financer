package v1

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"

	"github.com/ysomad/financer/internal/auth"
	pb "github.com/ysomad/financer/internal/gen/proto/expense/v1"
	pbconnect "github.com/ysomad/financer/internal/gen/proto/expense/v1/expensev1connect"
	"github.com/ysomad/financer/internal/postgres"
)

var _ pbconnect.CategoryServiceHandler = &CategoryServer{}

type CategoryServer struct {
	pbconnect.UnimplementedCategoryServiceHandler // TODO: REMOVE WHEN IMPLEMENT
	category                                      postgres.CategoryStorage
}

func NewCategoryServer(s postgres.CategoryStorage) *CategoryServer {
	return &CategoryServer{category: s}
}

func (s *CategoryServer) ListCategories(ctx context.Context, r *connect.Request[pb.ListCategoriesRequest]) (*connect.Response[pb.ListCategoriesResponse], error) {
	// handle not specified category type
	categoryType := ""
	if r.Msg.Type != 0 {
		categoryType = pb.CategoryType_name[int32(r.Msg.Type)]
	}

	list, err := s.category.GetAll(ctx, postgres.GetAllCategoriesParams{
		IdentityID:   auth.IdentityID(ctx),
		SearchQuery:  r.Msg.SearchQuery,
		CategoryType: categoryType,
		PageSize:     int(r.Msg.PageSize),
		PageToken:    r.Msg.PageToken,
	})
	if err != nil {
		slog.Error("categories not listed", "endpoint", r.Spec().Procedure, "err", err.Error())
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	cats := make([]*pb.Category, len(list.Categories))

	for i, cat := range list.Categories {
		cats[i] = &pb.Category{
			Name:   cat.Name,
			Type:   pb.CategoryType(pb.CategoryType_value[cat.Type]),
			Author: cat.Author.String,
		}
	}

	return connect.NewResponse(&pb.ListCategoriesResponse{
		Categories:    cats,
		NextPageToken: list.NextPageToken,
	}), nil
}

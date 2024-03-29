package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ysomad/financer/internal/paging"
	"github.com/ysomad/financer/internal/postgres/pgclient"
)

type CategoryStorage struct {
	*pgclient.Client
}

func NewCategoryStorage(c *pgclient.Client) CategoryStorage {
	return CategoryStorage{c}
}

type GetAllCategoriesParams struct {
	IdentityID   string
	SearchQuery  string
	CategoryType string
	PageSize     int32
	PageToken    string
}

type Category struct {
	Name      string      `db:"name"`
	Type      string      `db:"type"`
	Author    pgtype.Text `db:"author"`
	CreatedAt time.Time   `db:"created_at"`
}

type CategoryList struct {
	Categories    []Category
	NextPageToken string
}

func (s CategoryStorage) GetAll(ctx context.Context, p GetAllCategoriesParams) (CategoryList, error) {
	b := s.Builder.
		Select("c.name name, c.type type, c.author author, c.created_at created_at").
		From("identity_categories ic").
		InnerJoin("categories c ON ic.category = c.name").
		Where(sq.Eq{"ic.identity_id": p.IdentityID}).
		Where(sq.Eq{"c.deleted_at": nil}).
		OrderBy("c.created_at", "c.name").
		Limit(uint64(p.PageSize) + 1)

	if p.CategoryType != "" {
		b = b.Where(sq.Eq{"c.type": p.CategoryType})
	}

	if p.SearchQuery != "" {
		// fulltext search using PGroonga https://pgroonga.github.io
		b = b.Where(sq.Expr("c.name &@ ?", p.SearchQuery))
	}

	if p.PageToken != "" {
		prevName, prevTime, err := paging.Token(p.PageToken).Decode()
		if err != nil {
			return CategoryList{}, fmt.Errorf("page token not decoded: %w", err)
		}

		b = b.Where(sq.And{
			sq.GtOrEq{"c.created_at": prevTime},
			sq.GtOrEq{"c.name": prevName},
		})
	}

	sql, args, err := b.ToSql()
	if err != nil {
		return CategoryList{}, err
	}

	slog.Debug("list categorues query", "q", sql)

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return CategoryList{}, err
	}

	cats, err := pgx.CollectRows(rows, pgx.RowToStructByName[Category])
	if err != nil {
		return CategoryList{}, fmt.Errorf("rows not collected: %w", err)
	}

	list := CategoryList{Categories: cats}
	totalCats := len(cats)

	// has next page
	if totalCats == int(p.PageSize)+1 {
		list.Categories = cats[:totalCats-1]
		list.NextPageToken = paging.NewToken(cats[totalCats-1].Name, cats[totalCats-1].CreatedAt).String()
	}

	return list, nil
}

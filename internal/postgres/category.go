package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ysomad/financer/internal/domain"
	"github.com/ysomad/financer/internal/postgres/pgclient"
)

type CategoryStorage struct {
	*pgclient.Client
}

type category struct {
	ID     string      `db:"id"`
	Name   string      `db:"name"`
	Type   string      `db:"type"`
	Author pgtype.Int8 `db:"author"`
}

func (s CategoryStorage) ListByUserID(ctx context.Context, uid int64, catType domain.CatType) ([]category, error) {
	b := s.Builder.
		Select("c.id id, c.name name, c.type type, c.author author").
		From("user_categories uc").
		InnerJoin("categories c ON uc.category_id = c.id").
		Where(sq.Eq{"uc.user_id": uid}).
		Where(sq.Eq{"c.deleted_at": nil})

	if catType != domain.CatTypeUnspecified {
		b = b.Where(sq.Eq{"c.type": catType})
	}

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	cats, err := pgx.CollectRows(rows, pgx.RowToStructByName[category])
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return cats, nil
}

func (s CategoryStorage) FindByID(ctx context.Context, catID string) (category, error) {
	sql, args, err := s.Builder.
		Select("id, name, type, author").
		From("categories").
		Where(sq.Eq{"id": catID}).
		ToSql()
	if err != nil {
		return category{}, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return category{}, fmt.Errorf("query: %w", err)
	}

	cat, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category])
	if err != nil {
		return category{}, fmt.Errorf("scan: %w", err)
	}

	return cat, nil
}

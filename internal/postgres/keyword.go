package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/ysomad/financer/internal/domain"
	"github.com/ysomad/financer/internal/postgres/pgclient"
)

type KeywordStorage struct {
	*pgclient.Client
}

func (s KeywordStorage) FindCategory(ctx context.Context, uid int64, opName string, ct domain.CatType) (category, error) {
	sql, args, err := s.Builder.
		Select("c.id id, c.name name, c.author author, c.type type").
		From("user_keywords uk").
		InnerJoin("categories c ON uk.category_id = c.id").
		Where(sq.Eq{"uk.user_id": uid}).
		Where(sq.Eq{"uk.operation": opName}).
		Where(sq.Eq{"c.type": ct}).
		ToSql()
	if err != nil {
		return category{}, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return category{}, fmt.Errorf("query: %w", err)
	}

	cat, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[category])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return category{}, ErrNotFound
		}

		return category{}, fmt.Errorf("scan: %w", err)
	}

	return cat, nil
}

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

func (s *KeywordStorage) FindCategory(ctx context.Context, uid int64, opName string, ct domain.CatType) (Category, error) {
	sql, args, err := s.Builder.
		Select("c.id id, c.name name, c.author author, c.type type").
		From("user_keywords uk").
		InnerJoin("categories c ON uk.category_id = c.id").
		Where(sq.Eq{"uk.user_id": uid}).
		Where(sq.Eq{"uk.operation": opName}).
		Where(sq.Eq{"c.type": ct}).
		ToSql()
	if err != nil {
		return Category{}, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return Category{}, fmt.Errorf("query: %w", err)
	}

	cat, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Category])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrNotFound
		}

		return Category{}, fmt.Errorf("scan: %w", err)
	}

	return cat, nil
}

func (s *KeywordStorage) DeleteAll(ctx context.Context, uid int64) error {
	sql, args, err := s.Builder.
		Delete("user_keywords").
		Where(sq.Eq{"user_id": uid}).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := s.Pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

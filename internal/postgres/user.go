package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/ysomad/financer/internal/domain"
	"github.com/ysomad/financer/internal/postgres/pgclient"
)

type UserStorage struct {
	*pgclient.Client
}

type user struct {
	ID       int64  `db:"id"`
	Currency string `db:"currency"`
	Language string `db:"language"`
}

type CreateUserParams struct {
	UID       int64
	Currency  string
	Language  string
	CreatedAt time.Time
}

func (s *UserStorage) Create(ctx context.Context, p CreateUserParams) error {
	sql1, args1, err := s.Builder.
		Insert("users").
		Columns("id, currency, language, created_at").
		Values(p.UID, p.Currency, p.Language, p.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}

	sql2, args2, err := s.Builder.
		Insert("user_categories").
		Columns("user_id, category_id").
		Select(
			sq.Select(fmt.Sprintf("'%d'", p.UID), "id").
				From("categories").
				Where(sq.Eq{"author": nil}),
		).
		ToSql()
	if err != nil {
		return err
	}

	return pgx.BeginTxFunc(ctx, s.Pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sql1, args1...); err != nil {
			return fmt.Errorf("user not created: %w", err)
		}

		if _, err := tx.Exec(ctx, sql2, args2...); err != nil {
			return fmt.Errorf("categories not attached: %w", err)
		}

		return nil
	})
}

func (s *UserStorage) Find(ctx context.Context, uid int64) (domain.User, error) {
	sql, args, err := s.Builder.
		Select("id, currency, language").
		From("users").
		Where(sq.Eq{"id": uid}).
		ToSql()
	if err != nil {
		return domain.User{}, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return domain.User{}, fmt.Errorf("query: %w", err)
	}

	res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}

		return domain.User{}, fmt.Errorf("scan: %w", err)
	}

	return domain.User(res), nil
}

type UpdateParams struct {
	UID       int64
	Language  string
	Currency  string
	UpdatedAt time.Time
}

func (s *UserStorage) Update(ctx context.Context, p UpdateParams) error {
	sql, args, err := s.Builder.
		Update("users").
		Set("language", p.Language).
		Set("currency", p.Currency).
		Set("updated_at", p.UpdatedAt).
		Where(sq.Eq{"id": p.UID}).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := s.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

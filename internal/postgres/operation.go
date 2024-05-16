package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ysomad/financer/internal/money"
	"github.com/ysomad/financer/internal/postgres/pgclient"
)

type OperationStorage struct {
	*pgclient.Client
}

type SaveOperationParams struct {
	ID        string
	UID       int64
	CatID     string
	Operation string
	Currency  string
	Money     money.Money
	OccuredAt time.Time
	CreatedAt time.Time
}

func (s OperationStorage) Save(ctx context.Context, p SaveOperationParams) error {
	sql1, args1, err := s.Builder.
		Insert("operations").
		Columns("id, user_id, category_id, name",
			"currency, money, occured_at, created_at").
		Values(p.ID, p.UID, p.CatID, p.Operation,
			p.Currency, p.Money, p.OccuredAt, p.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}

	sql2, args2, err := s.Builder.
		Insert("user_keywords").
		Columns("user_id, category_id, operation").
		Values(p.UID, p.CatID, p.Operation).
		ToSql()
	if err != nil {
		return err
	}

	err = pgx.BeginTxFunc(ctx, s.Pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if _, err := s.Pool.Exec(ctx, sql1, args1...); err != nil {
			return fmt.Errorf("operation not saved: %w", err)
		}

		if _, err := s.Pool.Exec(ctx, sql2, args2...); err != nil {
			var pgErr *pgconn.PgError

			// do not save same keyword again
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return nil
			}

			return fmt.Errorf("keywords not saved: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}

	return nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ysomad/financer/internal/postgres/pgclient"
)

type ExpenseStorage struct {
	*pgclient.Client
}

func NewExpenseStorage(c *pgclient.Client) ExpenseStorage {
	return ExpenseStorage{
		Client: c,
	}
}

type SaveExpenseParams struct {
	ID         string
	IdentityID string
	Category   string
	Name       string
	Currency   string
	MoneyUnits int64
	MoneyNanos int32
	Date       time.Time
	CreatedAt  time.Time
}

type Expense struct {
	ID         string             `db:"id"`
	IdentityID string             `db:"identity_id"`
	Category   string             `db:"category"`
	Name       string             `db:"name"`
	Currency   string             `db:"currency"`
	MoneyUnits int64              `db:"money_units"`
	MoneyNanos int32              `db:"money_nanos"`
	Date       time.Time          `db:"date"`
	CreatedAt  time.Time          `db:"created_at"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at"`
	DeletedAt  pgtype.Timestamptz `db:"deleted_at"`
}

func (s ExpenseStorage) Save(ctx context.Context, p SaveExpenseParams) error {
	sql, args, err := s.Builder.
		Insert("expenses").
		Columns(
			"id",
			"identity_id",
			"category",
			"name",
			"currency",
			"money_units",
			"money_nanos",
			"date",
			"created_at").
		Values(
			p.ID,
			p.IdentityID,
			p.Category,
			p.Name,
			p.Currency,
			p.MoneyUnits,
			p.MoneyNanos,
			p.Date,
			p.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := s.Pool.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	return nil
}

func (s ExpenseStorage) Find(ctx context.Context, identityID, expense string) (Expense, error) {
	sql, args, err := s.Builder.Select(
		"id",
		"identity_id",
		"category",
		"name",
		"currency",
		"money_units",
		"money_nanos",
		"date",
		"created_at").
		From("expenses").
		Where(sq.Eq{"name": expense, "identity_id": identityID}).
		ToSql()
	if err != nil {
		return Expense{}, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return Expense{}, fmt.Errorf("query failed: %w", err)
	}

	e, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Expense])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Expense{}, ErrNotFound
		}

		return Expense{}, fmt.Errorf("expense not collected from rows: %w", err)
	}

	return e, nil
}

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

type IdentityStorage struct {
	*pgclient.Client
}

func NewIdentityStorage(c *pgclient.Client) *IdentityStorage {
	return &IdentityStorage{
		Client: c,
	}
}

var ErrIdentityNotFound = errors.New("identity not found")

type Identity struct {
	ID              string             `db:"id"`
	CreatedAt       time.Time          `db:"created_at"`
	UpdatedAt       pgtype.Timestamptz `db:"updated_at"`
	DeletedAt       pgtype.Timestamptz `db:"deleted_at"`
	TelegramUID     pgtype.Int8        `db:"telegram_uid"`
	DefaultCurrency string             `db:"default_currency"`
}

type InsertIdentityIn struct {
	ID              string
	CreatedAt       time.Time
	TelegramUID     int64
	DefaultCurrency string
}

func (s *IdentityStorage) Insert(ctx context.Context, in InsertIdentityIn) error {
	sql1, args1, err := s.Builder.
		Insert("identities").
		Columns("id, created_at").
		Values(in.ID, in.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}

	sql2, args2, err := s.Builder.
		Insert("identity_traits").
		Columns("telegram_uid, identity_id, default_currency").
		Values(in.TelegramUID, in.ID, in.DefaultCurrency).
		ToSql()
	if err != nil {
		return err
	}

	// add categories without author to identity (default categories)
	sql3, args3, err := s.Builder.
		Insert("identity_categories").
		Columns("identity_id", "category").
		Select(
			sq.Select(fmt.Sprintf("'%s'", in.ID), "name").
				From("categories").
				Where(sq.Eq{"author": nil})).ToSql()
	if err != nil {
		return err
	}

	err = pgx.BeginTxFunc(ctx, s.Pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, sql1, args1...); err != nil {
			return fmt.Errorf("identity not saved: %w", err)
		}

		if _, err := tx.Exec(ctx, sql2, args2...); err != nil {
			return fmt.Errorf("identity traits not saved: %w", err)
		}

		if _, err := tx.Exec(ctx, sql3, args3...); err != nil {
			return fmt.Errorf("default categories not attached to identity: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tx not commited: %w", err)
	}

	return nil
}

func (s *IdentityStorage) FindByTelegramUID(ctx context.Context, uid int64) (Identity, error) {
	sql, args, err := s.Builder.
		Select("i.id id",
			"i.created_at created_at",
			"i.updated_at updated_at",
			"i.deleted_at deleted_at",
			"t.default_currency default_currency",
			"t.telegram_uid telegram_uid").
		From("identity_traits t").
		InnerJoin("identities i on t.identity_id = i.id").
		Where(sq.Eq{"telegram_uid": uid}).ToSql()
	if err != nil {
		return Identity{}, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return Identity{}, fmt.Errorf("error fetching identity: %w", err)
	}

	id, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Identity])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Identity{}, ErrIdentityNotFound
		}

		return Identity{}, fmt.Errorf("error getting result from row: %w", err)
	}

	return id, nil
}

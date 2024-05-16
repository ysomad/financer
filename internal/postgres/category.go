package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ysomad/financer/internal/domain"
	"github.com/ysomad/financer/internal/postgres/pgclient"
)

type CategoryStorage struct {
	*pgclient.Client
}

type Category struct {
	ID     string         `db:"id"`
	Name   string         `db:"name"`
	Type   domain.CatType `db:"type"`
	Author pgtype.Int8    `db:"author"`
}

func (s CategoryStorage) ListByUserID(ctx context.Context, uid int64, catType domain.CatType) ([]Category, error) {
	b := s.Builder.
		Select("c.id id, c.name name, c.type type, c.author author").
		From("user_categories uc").
		InnerJoin("categories c ON uc.category_id = c.id").
		Where(sq.And{
			sq.Eq{"uc.user_id": uid},
			sq.Eq{"c.deleted_at": nil},
			sq.NotEq{"c.type": domain.CatTypeOther}, // never return OTHER category
		})

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

	cats, err := pgx.CollectRows(rows, pgx.RowToStructByName[Category])
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return cats, nil
}

func (s CategoryStorage) FindByID(ctx context.Context, catID string) (Category, error) {
	sql, args, err := s.Builder.
		Select("id, name, type, author").
		From("categories").
		Where(sq.Eq{"id": catID}).
		ToSql()
	if err != nil {
		return Category{}, err
	}

	rows, err := s.Pool.Query(ctx, sql, args...)
	if err != nil {
		return Category{}, fmt.Errorf("query: %w", err)
	}

	cat, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Category])
	if err != nil {
		return Category{}, fmt.Errorf("scan: %w", err)
	}

	return cat, nil
}

type SaveCategoryParams struct {
	ID        string
	Name      string
	Type      domain.CatType
	Author    int64
	CreatedAt time.Time
}

func (s CategoryStorage) Save(ctx context.Context, p SaveCategoryParams) error {
	sql, args, err := s.Builder.
		Insert("categories").
		Columns("id, name, type, author, created_at").
		Values(p.ID, p.Name, p.Type, p.Author, p.CreatedAt).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := s.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

type ReplaceCategoryParams struct {
	UID   int64
	OldID string
	NewID string
}

// Replace replaces category from OldID to NewID in user categories.
func (s CategoryStorage) Replace(ctx context.Context, p ReplaceCategoryParams) error {
	sql, args, err := s.Builder.
		Update("user_categories").
		Set("category_id", p.NewID).
		Where(sq.And{
			sq.Eq{"user_id": p.UID},
			sq.Eq{"category_id": p.OldID},
		}).
		ToSql()
	if err != nil {
		return err
	}

	if _, err := s.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

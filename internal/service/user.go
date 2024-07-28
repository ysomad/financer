package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ysomad/financer/internal/domain"
	"github.com/ysomad/financer/internal/postgres"
)

const (
	defaultCurrency = "USD"
	defaultLanguage = "en"
)

type User struct {
	storage *postgres.UserStorage
}

func NewUser(s *postgres.UserStorage) *User {
	return &User{storage: s}
}

func (u *User) GetOrCreate(ctx context.Context, uid int64) (domain.User, error) {
	usr, err := u.storage.Find(ctx, uid)
	if err == nil {
		return usr, nil
	}

	if !errors.Is(err, postgres.ErrNotFound) {
		return domain.User{}, fmt.Errorf("repository failed: %w", err)
	}

	slog.InfoContext(ctx, "creating new user")

	params := postgres.CreateUserParams{
		UID:       uid,
		Currency:  defaultCurrency,
		Language:  defaultLanguage,
		CreatedAt: time.Now(),
	}

	if err = u.storage.Create(ctx, params); err != nil {
		return domain.User{}, fmt.Errorf("user not created: %w", err)
	}

	return domain.User{
		ID:       params.UID,
		Currency: params.Currency,
		Language: params.Language,
	}, nil
}

func (u *User) Update(ctx context.Context, usr domain.User) error {
	if err := usr.Validate(); err != nil {
		return fmt.Errorf("user not valid before update: %w", err)
	}

	if err := u.storage.Update(ctx, postgres.UpdateParams{
		UID:       usr.ID,
		Language:  usr.Language,
		Currency:  usr.Currency,
		UpdatedAt: time.Now(),
	}); err != nil {
		return fmt.Errorf("user not updated: %w", err)
	}

	return nil
}

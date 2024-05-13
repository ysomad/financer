package domain

import (
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/rmg/iso4217"
)

var supportedLanguages = []string{"en", "ru"}

var (
	ErrUnsupportedLanguage = errors.New("unsupported language")
	ErrUnsupportedCurrency = errors.New("unsupported currency")
)

type User struct {
	ID       int64
	Currency string
	Language string
}

func (u *User) Validate() error {
	u.Language = strings.ToLower(u.Language)
	u.Currency = strings.ToUpper(u.Currency)

	if !slices.Contains(supportedLanguages, u.Language) {
		return ErrUnsupportedLanguage
	}

	code, _ := iso4217.ByName(u.Currency)
	if code == 0 {
		slog.Info(u.Currency, code)
		return ErrUnsupportedCurrency
	}

	return nil
}

func (u *User) IDString() string {
	return strconv.FormatInt(u.ID, 10)
}

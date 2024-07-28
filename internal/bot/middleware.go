package bot

import (
	"context"
	"fmt"

	tele "gopkg.in/telebot.v3"

	"github.com/ysomad/financer/internal/domain"
)

const (
	stdCtxKey  = "stdcontext"
	userCtxKey = "myepicuser"
)

// stdContext returns context.Context from telebot context
func stdContext(c tele.Context) context.Context {
	v := c.Get(stdCtxKey)
	if v == nil {
		return context.Background()
	}
	if ctx, ok := v.(context.Context); ok {
		return ctx
	}
	return context.Background()
}

func contextMiddleware(version string) func(next tele.HandlerFunc) tele.HandlerFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			ctx := context.Background()
			ctx = withRecipient(ctx, c.Recipient().Recipient())
			ctx = withVersion(ctx, version)
			c.Set(stdCtxKey, ctx)
			return next(c)
		}
	}
}

// userContextMiddleware gets or creates user from db and sets it to context.
func (b *Bot) userContextMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		usr, err := b.user.GetOrCreate(stdContext(c), c.Chat().ID)
		if err != nil {
			return fmt.Errorf("user context: %w", err)
		}
		c.Set(userCtxKey, usr)
		return next(c)
	}
}

func userFromContext(c tele.Context) (domain.User, bool) {
	v := c.Get(userCtxKey)
	usr, ok := v.(domain.User)

	if v == nil || !ok || usr.ID == 0 || usr.Language == "" || usr.Currency == "" {
		return domain.User{}, false
	}

	return usr, true
}

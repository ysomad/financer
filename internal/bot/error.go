package bot

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/ysomad/financer/internal/bot/msg"

	tele "gopkg.in/telebot.v3"
)

var (
	errUserNotInContext        = errors.New("user not found in context")
	errStateNotFound           = errors.New("no state found")
	errInvalidStateData        = errors.New("invalid state data")
	errUnsupportedCallbackData = errors.New("unsupported callback data")
)

func (b *Bot) HandleError(err error, c tele.Context) {
	ctx := stdContext(c)

	if c == nil {
		slog.WarnContext(ctx, "empty telebot context in error handler")
		return
	}

	slog.ErrorContext(ctx, err.Error())

	usr, ok := userFromContext(c)
	if !ok {
		slog.WarnContext(ctx, "no user in context while handling error")
		usr.Language = defaultLang
	}

	if err := c.Send(msg.Get(msg.InternalError, usr.Language)); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("error msg not sent to user: %s", err.Error()))
	}
}

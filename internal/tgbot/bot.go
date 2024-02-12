package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/ladydascalie/currency"
	goredis "github.com/redis/go-redis/v9"
	"gopkg.in/telebot.v3"

	expensev1 "github.com/ysomad/financer/internal/gen/proto/expense/v1"
	"github.com/ysomad/financer/internal/gen/proto/expense/v1/expensev1connect"
	pb "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	"github.com/ysomad/financer/internal/gen/proto/telegram/v1/telegramv1connect"
	"github.com/ysomad/financer/internal/tgbot/config"
	"github.com/ysomad/financer/internal/tgbot/model"
	"github.com/ysomad/financer/internal/tgbot/redis"
)

var messages = map[string]string{
	"internal_error":         "I'm not feeling good right now, try later...",
	"currency_selection":     "Choose from list or send any other currency in ISO-4217 format (for example UAH, KZT, GBP etc):",
	"currency_set":           "%s saved as your default currency. Next time you send me a command without specifying currency I'll use %s.\n\nYou can always change default currency by using /set_currency command",
	"invalid_currency":       "Please provide currency code in ISO-4217 format...",
	"canceled":               "Current operation is canceled",
	"invalid_expense_format": "Expense must be in format: {?+}{money amount} {expense} {?currency} {?date in format 20.05 or 20.05.1999}",
	"invalid_date":           "Date must be in format: '20.01.2006' or '20.01.06' or '20.01'",
}

type IdentityService struct {
	Client telegramv1connect.IdentityServiceClient
	Cache  redis.IdentityCache
}

type Bot struct {
	conf        config.Config
	redis       *goredis.Client
	identity    IdentityService
	accessToken telegramv1connect.AccessTokenServiceClient
	category    expensev1connect.CategoryServiceClient
	state       redis.StateCache
	expense     expensev1connect.ExpenseServiceClient
}

func New(
	conf config.Config,
	rdb *goredis.Client,
	id IdentityService,
	accessToken telegramv1connect.AccessTokenServiceClient,
	category expensev1connect.CategoryServiceClient,
	state redis.StateCache,
	expense expensev1connect.ExpenseServiceClient,
) *Bot {
	return &Bot{
		conf:        conf,
		identity:    id,
		accessToken: accessToken,
		category:    category,
		redis:       rdb,
		state:       state,
		expense:     expense,
	}
}

func accessTokenKey(tgUID int64) string {
	return fmt.Sprintf("access_token:%d", tgUID)
}

func (b *Bot) Start(c telebot.Context) error {
	if _, err := b.authorize(context.Background(), c.Chat().ID); err != nil {
		slog.Error("/start not authorized", "err", err)
		return c.Send(messages["internal_error"])
	}

	/* send user instruction with menu
	   - /add {amount money} {?iso 4217 currency} {comment} {date in format 20.05 or 20.05.1999} - adds expense
	*/

	return nil
}

func (b *Bot) CmdCategories(c telebot.Context) error {
	tguid := c.Chat().ID
	ctx := context.Background()

	identity, err := b.authorize(ctx, tguid)
	if err != nil {
		slog.Error("/categories not authorized", "err", err)
		return c.Send(messages["internal_error"])
	}

	resp, err := b.category.ListCategories(ctx, withAccessToken(&expensev1.ListCategoriesRequest{
		PageSize: 50,
	}, identity.AccessToken))
	if err != nil {
		slog.Error("categories not listed", "err", err.Error())
		return c.Send(err.Error())
	}

	sb := strings.Builder{}
	sb.Grow(len(resp.Msg.Categories) + 2)
	sb.WriteString("➖ Expenses:\n\n")

	// Expenses
	for _, cat := range resp.Msg.Categories {
		if cat.Type != 1 { // TODO: refactor
			continue
		}

		if _, err := sb.WriteString(cat.Name + "\n"); err != nil {
			slog.Error("category not writed to builder", "err", err.Error())
			return c.Send(messages["internal_error"])
		}
	}

	sb.WriteString("\n➕ Earnings:\n\n")

	// Earnings
	for _, cat := range resp.Msg.Categories {
		if cat.Type != 2 { // TODO: refactor
			continue
		}

		if _, err := sb.WriteString(cat.Name + "\n"); err != nil {
			slog.Error("category not writed to builder", "err", err.Error())
			return c.Send(messages["internal_error"])
		}
	}

	return c.Send(sb.String())
}

func (b *Bot) CmdSetCurrency(c telebot.Context) error {
	tguid := c.Chat().ID

	// TODO: with timeout
	ctx := context.Background()

	_, err := b.authorize(ctx, tguid)
	if err != nil {
		slog.Error("/set_currency not authorized", "err", err)
		return c.Send(messages["internal_error"])
	}

	kb := new(telebot.ReplyMarkup)

	btnRUB := kb.Data("🇷🇺 Rubles", "set_currency", "RUB")
	btnUSD := kb.Data("🇺🇸 Dollars", "set_currency", "USD")
	btnEUR := kb.Data("🇪🇺 Euros", "set_currency", "EUR")
	btnCancel := kb.Data("Cancel", "cancel")

	kb.Inline(
		kb.Row(btnUSD),
		kb.Row(btnRUB),
		kb.Row(btnEUR),
		kb.Row(btnCancel))

	if err := b.state.Save(ctx, tguid, model.StateCurrencySelection); err != nil {
		slog.Error("state not saved", "err", err.Error())
		return c.Send(messages["internal_error"])
	}

	return c.Send(messages["currency_selection"], kb)
}

var errInvalidCurrencyCode = errors.New("currency code must be in iso-4217 format")

// saveCurrency saves new currency to server
func (b *Bot) saveCurrency(ctx context.Context, tguid int64, currCode string) error {
	// TODO: test
	currCode = strings.ToUpper(currCode)

	if !currency.Valid(currCode) {
		return errInvalidCurrencyCode
	}

	identity, err := b.identity.Cache.Get(ctx, tguid)
	if err != nil {
		return fmt.Errorf("identity not found in cache")
	}

	if _, err := b.identity.Client.UpdateIdentity(ctx, withAPIKey(&pb.UpdateIdentityRequest{
		Id:       identity.ID,
		Currency: currCode,
	}, b.conf.Server.APIKey)); err != nil {
		return fmt.Errorf("identity not updated: %w", err)
	}

	return nil
}

func msgCurrencySet(currency string) string {
	currency = strings.ToUpper(currency)
	return fmt.Sprintf(messages["currency_set"], currency, currency)
}

func (b *Bot) HandleText(c telebot.Context) error {
	tguid := c.Chat().ID
	ctx := context.Background()

	state, err := b.state.Get(ctx, tguid)
	if err != nil {
		slog.Error("couldnt get state", "err", err.Error())
		return c.Send(messages["internal_error"])
	}

	switch state {
	case model.StateCurrencySelection:
		currency := c.Text()

		if err := b.saveCurrency(ctx, tguid, c.Text()); err != nil {
			if errors.Is(err, errInvalidCurrencyCode) {
				return c.Send(messages["invalid_currency"])
			}

			return c.Send(messages["internal_error"])
		}

		if err := b.state.Del(ctx, tguid); err != nil {
			slog.Error("state not deleted", "err", err.Error())
		}

		return c.Send(msgCurrencySet(currency))
	default:
		_, err := b.authorize(ctx, tguid)
		if err != nil {
			slog.Error("identity not found in cache", err)
			return c.Send(messages["internal_error"])
		}

		// handle expense creation
		//`{?+}{money amount} {expense} {?currency} {?date in format 20.05 or 20.05.1999}`:
		args := strings.Split(c.Text(), " ")
		argsNum := len(args)

		slog.Debug("ARGS", "ARGS", args)

		if argsNum < 2 {
			return c.Send(messages["invalid_expense_format"])
		}

		catType := expensev1.CategoryType_EXPENSES
		money := strings.ReplaceAll(args[0], "-", "")

		if strings.HasPrefix(money, "+") {
			catType = expensev1.CategoryType_EARNINGS
			money = strings.ReplaceAll(money, "+", "")
		}

		expense := args[1]
		curr := ""
		date := time.Now()

		if argsNum == 3 {
			arg2 := strings.ToUpper(args[2]) // currency or date

			if currency.Valid(arg2) {
				curr = arg2
				date = time.Now()
			} else {
				date, err = parseDate(arg2)
				if err != nil {
					slog.Error("date not parsed", "err", err.Error())
					return c.Send(messages["invalid_date"])
				}
			}
		}

		if argsNum == 4 {
			curr = strings.ToUpper(args[2])
			if !currency.Valid(curr) {
				return c.Send(messages["invalid_currency"])
			}

			date, err = parseDate(args[3])
			if err != nil {
				return c.Send(messages["invalid_date"])
			}
		}

		slog.Debug("EXPENSE", "cat_type", catType, "exp", expense, "curr", curr, "year", date.Year(), "month", date.Month(), "day", date.Day(), "money", money)
	}

	return nil
}

func parseDate(s string) (time.Time, error) {
	var (
		formats = [3]string{"02.01.2006", "02.01.06", "02.01"}
		t       time.Time
		err     error
	)

	for _, layout := range formats {
		t, err = time.Parse(layout, s)
		if err == nil {
			if t.Year() == 0 {
				t = time.Date(time.Now().Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			}

			return t, nil
		}
	}

	return t, fmt.Errorf("invalid date format: %s", s)
}

func (b *Bot) HandleCallback(c telebot.Context) error {
	// TODO: with timeout
	ctx := context.Background()

	cb := c.Callback()
	tguid := c.Chat().ID

	slog.Debug("callback", "data", cb.Data, "unique", cb.Unique)

	cbData := strings.ReplaceAll(cb.Data, "\f", "")
	cbDataParts := strings.Split(cbData, "|")

	switch len(cbDataParts) {
	case 1: // callback without data, only unique preset
		if cbDataParts[0] == "cancel" {
			if err := b.state.Del(ctx, tguid); err != nil {
				slog.Error("state not deleted on cancel", "err", err.Error())
			}

			return c.Edit(messages["canceled"])
		}
	case 2: // callback with unique and data
		if cbDataParts[0] == "set_currency" {
			currency := cbDataParts[1]

			if err := b.saveCurrency(ctx, tguid, currency); err != nil {
				slog.Error("currency not saved", "err", err.Error())
				return c.Send(messages["internal_error"])
			}

			if err := b.state.Del(ctx, tguid); err != nil {
				slog.Error("state not deleted on /set_currency", "err", err.Error())
			}

			return c.Edit(msgCurrencySet(currency))
		}
	default:
		slog.Error("unsupported callback", "data", cb.Data)
		return nil
	}

	return nil
}

// getOrCreateIdentity gets or creates identity from server.
func (b *Bot) getOrCreateIdentity(ctx context.Context, tguid int64) (*pb.Identity, error) {
	resp, err := b.identity.Client.GetIdentity(ctx, withAPIKey(&pb.GetIdentityRequest{
		TgUid: tguid,
	}, b.conf.Server.APIKey))
	if err == nil {
		return resp.Msg, nil
	}

	if connectErr := new(connect.Error); errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
		resp, err := b.identity.Client.CreateIdentity(ctx, withAPIKey(&pb.CreateIdentityRequest{
			TgUid: tguid,
		}, b.conf.Server.APIKey))
		if err != nil {
			return nil, fmt.Errorf("identity not created: %w", err)
		}

		return resp.Msg, nil
	}

	// server error
	slog.Error("cannot get identity from server", "err", err.Error())

	return nil, err
}

// authorize returns identity from cache or creates it and issues access token.
func (b *Bot) authorize(ctx context.Context, tguid int64) (model.Identity, error) {
	// get identity from cache
	identity, err := b.identity.Cache.Get(ctx, tguid)
	if err == nil {
		return identity, nil
	}

	if !errors.Is(err, redis.ErrNotFound) {
		slog.Error("cache error getting identity", "err", err.Error())
	}

	slog.Info("identity not found in cache", "tg_uid", tguid)

	// get pbIdentity from server
	pbIdentity, err := b.getOrCreateIdentity(ctx, tguid)
	if err != nil {
		return model.Identity{}, fmt.Errorf("couldnt get identity from server: %w", err)
	}

	// issue access token for newly created identity
	resp, err := b.accessToken.IssueAccessToken(ctx, withAPIKey(&pb.IssueAccessTokenRequest{
		TgUid: tguid,
	}, b.conf.Server.APIKey))
	if err != nil {
		return model.Identity{}, fmt.Errorf("access token not issued: %w", err)
	}

	identity = model.Identity{
		ID:          pbIdentity.Id,
		TGUID:       pbIdentity.TgUid,
		AccessToken: resp.Msg.AccessToken,
	}

	if err := b.identity.Cache.Save(ctx, identity); err != nil {
		slog.Error("identity not saved to cache", "err", err.Error())
	}

	return identity, nil
}

func withAPIKey[T any](msg *T, apiKey string) *connect.Request[T] {
	r := connect.NewRequest(msg)
	r.Header().Set("X-API-KEY", apiKey)
	return r
}

func withAccessToken[T any](msg *T, accessToken string) *connect.Request[T] {
	r := connect.NewRequest(msg)
	// TODO: move to httponly secure cookie
	r.Header().Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	return r
}

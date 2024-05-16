package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	iso6391 "github.com/emvi/iso-639-1"
	"github.com/google/uuid"
	"github.com/ysomad/financer/internal/bot/msg"
	"github.com/ysomad/financer/internal/date"
	"github.com/ysomad/financer/internal/domain"
	"github.com/ysomad/financer/internal/money"
	"github.com/ysomad/financer/internal/postgres"
	"github.com/ysomad/financer/internal/service"

	"github.com/hashicorp/golang-lru/v2/expirable"
	tele "gopkg.in/telebot.v3"
)

const defaultLang = "en"

type Bot struct {
	state     *expirable.LRU[string, domain.State]
	category  postgres.CategoryStorage
	user      service.User
	operation postgres.OperationStorage
	keyword   postgres.KeywordStorage
}

func New(st *expirable.LRU[string, domain.State], cat postgres.CategoryStorage,
	usr service.User, op postgres.OperationStorage, kw postgres.KeywordStorage,
) *Bot {
	return &Bot{
		state:     st,
		category:  cat,
		user:      usr,
		operation: op,
		keyword:   kw,
	}
}

func btnCancel(kb *tele.ReplyMarkup, lang string) tele.Btn {
	return kb.Data(msg.Get(msg.BtnCancel, lang), domain.StepCancel.String())
}

func (b *Bot) Start(c tele.Context) error {
	/*
		1 Спросить валюту и язык по умолчанию
		2 Создать юзера
		3 Положить юзера в кэш
	*/

	/* send user instruction with menu
	   - /add {amount money} {?iso 4217 currency} {comment} {date in format 20.05 or 20.05.1999} - adds expense
	*/

	return nil
}

func (b *Bot) ListCategories(c tele.Context) error {
	ctx := stdContext(c)

	cats, err := b.category.ListByUserID(ctx, c.Chat().ID, domain.CatTypeUnspecified)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return fmt.Errorf("list user categories: %w", err)
	}

	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	sb := strings.Builder{}
	sb.Grow(len(cats) + 2)

	// TODO: ugly ass shit

	if _, err := sb.WriteString(fmt.Sprintf("%s\n\n", msg.Get(msg.ExpenseCategoriesTitle, usr.Language))); err != nil {
		return fmt.Errorf("write expense title: %w", err)
	}

	for _, cat := range cats {
		if cat.Type != domain.CatTypeExpenses {
			continue
		}
		if _, err := sb.WriteString(fmt.Sprintf("%s\n", cat.Name)); err != nil {
			return fmt.Errorf("write expense category: %w", err)
		}
	}

	if _, err := sb.WriteString(fmt.Sprintf("\n%s\n\n", msg.Get(msg.IncomeCategoriesTitle, usr.Language))); err != nil {
		return fmt.Errorf("write income title: %w", err)
	}

	for _, cat := range cats {
		if cat.Type != domain.CatTypeIncome {
			continue
		}
		if _, err := sb.WriteString(fmt.Sprintf("%s\n", cat.Name)); err != nil {
			return fmt.Errorf("write income category: %w", err)
		}
	}

	return c.Send(sb.String())
}

func (b *Bot) RenameCategory(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	kb := &tele.ReplyMarkup{}
	step := domain.StepCategoryTypeSelection

	btnIncome := kb.Data(msg.Get(msg.BtnIncome, usr.Language), step.String(), domain.CatTypeIncome.String())
	btnExpenses := kb.Data(msg.Get(msg.BtnExpenses, usr.Language), step.String(), domain.CatTypeExpenses.String())

	kb.Inline(
		kb.Row(btnIncome),
		kb.Row(btnExpenses),
		kb.Row(btnCancel(kb, usr.Language)),
	)

	b.state.Add(usr.IDString(), domain.State{Step: step})

	return c.Send(msg.Get(msg.CategoryTypeSelection, usr.Language), kb)
}

func (b *Bot) SetLanguage(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	kb := &tele.ReplyMarkup{}
	step := domain.StepLanguageSelection

	btnRus := kb.Data(msg.Get(msg.BtnRussian, usr.Language), step.String(), "ru")
	btnEng := kb.Data(msg.Get(msg.BtnEnglish, usr.Language), step.String(), "en")

	kb.Inline(
		kb.Row(btnRus),
		kb.Row(btnEng),
		kb.Row(btnCancel(kb, usr.Language)),
	)

	b.state.Add(usr.IDString(), domain.State{Step: step})

	return c.Send(msg.Get(msg.LanguageSelection, usr.Language), kb)
}

func (b *Bot) SetCurrency(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	kb := &tele.ReplyMarkup{}
	step := domain.StepCurrencySelection

	btnRUB := kb.Data(msg.Get(msg.BtnRUB, usr.Language), step.String(), "RUB")
	btnUSD := kb.Data(msg.Get(msg.BtnUSD, usr.Language), step.String(), "USD")
	btnEUR := kb.Data(msg.Get(msg.BtnEUR, usr.Language), step.String(), "EUR")

	kb.Inline(
		kb.Row(btnUSD),
		kb.Row(btnRUB),
		kb.Row(btnEUR),
		kb.Row(btnCancel(kb, usr.Language)),
	)

	b.state.Add(usr.IDString(), domain.State{Step: step})

	return c.Send(msg.Get(msg.CurrencySelection, usr.Language), kb)
}

func (b *Bot) HandleText(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	ctx := stdContext(c)

	state, ok := b.state.Get(usr.IDString())
	if !ok {
		slog.InfoContext(ctx, "no state in text handler")
	}

	switch state.Step {
	case domain.StepCurrencySelection:
		defer b.state.Remove(usr.IDString())

		usr, err := b.user.Update(ctx, domain.User{
			ID:       usr.ID,
			Currency: c.Text(),
			Language: usr.Language,
		})
		if err != nil {
			if errors.Is(err, domain.ErrUnsupportedCurrency) {
				return c.Send(msg.Get(msg.InvalidCurrency, usr.Language))
			}

			return fmt.Errorf("user not updated on text handle: %w", err)
		}

		return c.Send(msg.Getf(msg.CurrencySaved, usr.Language, usr.Currency, usr.Currency))
	case domain.StepCategoryRename:
		defer b.state.Remove(usr.IDString())

		cat, ok := state.Data.(postgres.Category)
		if !ok {
			return fmt.Errorf("category rename: %w", errInvalidStateData)
		}

		newCatName := c.Text()
		newCatID := uuid.NewString()

		// TODO: wrap into tx
		if err := b.category.Save(ctx, postgres.SaveCategoryParams{
			ID:        newCatID,
			Name:      newCatName,
			Type:      cat.Type,
			Author:    usr.ID,
			CreatedAt: time.Now(),
		}); err != nil {
			return fmt.Errorf("category not saved: %w", err)
		}

		if err := b.category.Replace(ctx, postgres.ReplaceCategoryParams{
			UID:   usr.ID,
			OldID: cat.ID,
			NewID: newCatID,
		}); err != nil {
			return fmt.Errorf("category not replaced: %w", err)
		}

		return c.Send(msg.Getf(msg.CategoryRenamed, usr.Language, cat.Name, newCatName))
	default:
		// handle operation save
		parts := strings.Split(c.Text(), " ")
		if len(parts) < 2 {
			return c.Send(msg.Get(msg.InvalidOperationFormat, usr.Language))
		}

		moneyStr := parts[0]

		// костыль
		if !strings.Contains(moneyStr, "-") && !strings.Contains(moneyStr, "+") {
			moneyStr = "-" + moneyStr
		}

		money, err := money.Parse(moneyStr)
		if err != nil {
			return c.Send(msg.Get(msg.InvalidOperationFormat, usr.Language))
		}

		if money == 0 {
			return c.Send(msg.Get(msg.InvalidOperationFormat, usr.Language))
		}

		occuredAt := time.Now()
		opName := parts[1]

		// parse date from last argument
		last := len(parts) - 1

		if len(parts) > 2 {
			tmpDate, err := date.Parse(parts[last])
			if err != nil {
				slog.InfoContext(ctx, "date not parsed", "input", parts[last])
				last++
			} else {
				occuredAt = tmpDate
			}

			opName = strings.Join(parts[1:last], " ")
		}

		catType := domain.CatTypeExpenses

		if money > 0 {
			catType = domain.CatTypeIncome
		}

		// find operation with the same name
		cat, err := b.keyword.FindCategory(ctx, usr.ID, opName, catType)
		if err == nil {
			err := b.operation.Save(ctx, postgres.SaveOperationParams{
				ID:        uuid.New().String(),
				UID:       usr.ID,
				CatID:     cat.ID,
				Operation: opName,
				Currency:  usr.Currency,
				Money:     money,
				OccuredAt: occuredAt,
				CreatedAt: time.Now(),
			})
			if err != nil {
				return fmt.Errorf("operation not saved: %w", err)
			}

			if cat.Type == domain.CatTypeIncome {
				return c.Send(msg.Getf(msg.IncomeSaved, usr.Language, money.String(), usr.Currency, cat.Name, opName))
			}

			return c.Send(msg.Getf(msg.ExpenseSaved, usr.Language, money.String(), usr.Currency, cat.Name, opName))
		}
		if !errors.Is(err, postgres.ErrNotFound) {
			return fmt.Errorf("keyword search failed: %w", err)
		}

		step := domain.StepCategorySelection

		kb, err := b.categoriesKeyboard(ctx, usr, step, domain.CatType(catType), true)
		if err != nil {
			return fmt.Errorf("")
		}

		b.state.Add(usr.IDString(), domain.State{
			Step: step,
			Data: operation{
				name:      opName,
				money:     money,
				occuredAt: occuredAt,
			},
		})

		return c.Send(msg.Get(msg.CategorySelection, usr.Language), kb)
	}
}

// categoriesKeyboard builds inline keyboard with categories.
func (b *Bot) categoriesKeyboard(ctx context.Context, usr domain.User, nextStep domain.Step, ct domain.CatType, other bool) (*tele.ReplyMarkup, error) {
	// Ask for category select (only if operation with the same name not found)
	cats, err := b.category.ListByUserID(ctx, usr.ID, ct)
	if err != nil {
		return nil, fmt.Errorf("list categories failed: %w", err)
	}

	kb := &tele.ReplyMarkup{}

	btnRows := make([]tele.Row, 0, len(cats)/2+2)

	var tmp tele.Btn

	for i, cat := range cats {
		btn := kb.Data(cat.Name, nextStep.String(), cat.ID)

		// maximum 5 buttons in one column, optimal for mobile devices
		if len(cats) <= 5 {
			btnRows = append(btnRows, kb.Row(btn))
			continue
		}

		if i%2 == 0 {
			tmp = btn
			continue
		}

		btnRows = append(btnRows, []tele.Btn{tmp, btn})
	}

	if other {
		btnOther := kb.Data(msg.Get(msg.BtnOther, usr.Language), nextStep.String(), domain.OtherCategoryID)
		btnRows = append(btnRows, kb.Row(btnOther))
	}

	btnRows = append(btnRows, kb.Row(btnCancel(kb, usr.Language)))

	kb.Inline(btnRows...)

	return kb, nil
}

type buttonCallback struct {
	unique string
	data   string
}

func parseCallback(data string) (buttonCallback, error) {
	data = strings.TrimPrefix(data, "\f")
	dataparts := strings.Split(data, "|")

	switch len(dataparts) {
	case 1:
		return buttonCallback{unique: dataparts[0]}, nil
	case 2:
		return buttonCallback{unique: dataparts[0], data: dataparts[1]}, nil
	default:
		return buttonCallback{}, errUnsupportedCallbackData
	}
}

type operation struct {
	name      string
	money     money.Money
	occuredAt time.Time
}

func (b *Bot) HandleCallback(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	telecb := c.Callback()

	cb, err := parseCallback(telecb.Data)
	if err != nil {
		return err
	}

	ctx := stdContext(c)

	slog.InfoContext(ctx, "callback received", "unique", cb.unique, "data", cb.data, "callback_id", telecb.ID)

	switch cb.unique {
	case domain.StepCategorySelection.String():
		state, ok := b.state.Get(usr.IDString())
		if !ok {
			return fmt.Errorf("currency selection callback: %w", errStateNotFound)
		}

		op, ok := state.Data.(operation)
		if !ok {
			return fmt.Errorf("currency selection callback: %w", errInvalidStateData)
		}

		if err := b.operation.Save(ctx, postgres.SaveOperationParams{
			ID:        uuid.New().String(),
			UID:       usr.ID,
			CatID:     cb.data,
			Operation: op.name,
			Currency:  usr.Currency,
			Money:     op.money,
			OccuredAt: op.occuredAt,
			CreatedAt: time.Now(),
		}); err != nil {
			return fmt.Errorf("currency selection callback: %w", err)
		}

		cat, err := b.category.FindByID(ctx, cb.data)
		if err != nil {
			return fmt.Errorf("category not found: %w", err)
		}

		if op.money > 0 {
			return c.Edit(msg.Getf(msg.IncomeSaved, usr.Language, op.money.String(), usr.Currency, cat.Name, op.name))
		}

		return c.Edit(msg.Getf(msg.ExpenseSaved, usr.Language, op.money.String(), usr.Currency, cat.Name, op.name))
	case domain.StepCategoryTypeSelection.String():
		kb, err := b.categoriesKeyboard(ctx, usr, domain.StepCategoryRenameSelection, domain.CatType(cb.data), false)
		if err != nil {
			return fmt.Errorf("categories keyboard in callback: %w", err)
		}

		b.state.Add(usr.IDString(), domain.State{Step: domain.StepCategoryRenameSelection})

		return c.Edit(msg.Get(msg.CategoryRenameSelection, usr.Language), kb)
	case domain.StepCategoryRenameSelection.String():
		cat, err := b.category.FindByID(ctx, cb.data)
		if err != nil {
			return fmt.Errorf("category not found on category rename: %w", err)
		}

		b.state.Add(usr.IDString(), domain.State{
			Step: domain.StepCategoryRename,
			Data: cat,
		})

		return c.Edit(msg.Getf(msg.CategoryRename, usr.Language, cat.Name))
	case domain.StepCurrencySelection.String():
		defer b.state.Remove(usr.IDString())

		usr, err := b.user.Update(ctx, domain.User{
			ID:       usr.ID,
			Currency: cb.data,
			Language: usr.Language,
		})
		if err != nil {
			if errors.Is(err, domain.ErrUnsupportedCurrency) {
				return c.Send(msg.Get(msg.InvalidCurrency, usr.Language))
			}

			return fmt.Errorf("currency not set in callback: %w", err)
		}

		return c.Edit(msg.Getf(msg.CurrencySaved, usr.Language, usr.Currency, usr.Currency))
	case domain.StepLanguageSelection.String():
		defer b.state.Remove(usr.IDString())

		usr, err := b.user.Update(ctx, domain.User{
			ID:       usr.ID,
			Currency: usr.Currency,
			Language: cb.data,
		})
		if err != nil {
			return fmt.Errorf("language not set in callback: %w", err)
		}

		return c.Edit(msg.Getf(msg.LanguageSaved, usr.Language, iso6391.NativeName(usr.Language)))
	case domain.StepCancel.String():
		b.state.Remove(usr.IDString())
		return c.Edit(msg.Get(msg.OperationCanceled, usr.Language))
	default:
		return fmt.Errorf("unsupported callback unique: %s", cb.unique)
	}
}

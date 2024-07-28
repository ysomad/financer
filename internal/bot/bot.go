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
	"github.com/hashicorp/golang-lru/v2/expirable"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"github.com/ysomad/financer/internal/bot/msg"
	botstate "github.com/ysomad/financer/internal/bot/state"
	"github.com/ysomad/financer/internal/config"
	"github.com/ysomad/financer/internal/date"
	"github.com/ysomad/financer/internal/domain"
	"github.com/ysomad/financer/internal/money"
	"github.com/ysomad/financer/internal/postgres"
	"github.com/ysomad/financer/internal/service"
)

const defaultLang = "en"

type Bot struct {
	tele      *tele.Bot
	state     *expirable.LRU[string, botstate.State]
	category  *postgres.CategoryStorage
	user      *service.User
	operation *postgres.OperationStorage
	keyword   *postgres.KeywordStorage
}

func New(conf config.Config, st *expirable.LRU[string, botstate.State], cat *postgres.CategoryStorage,
	usr *service.User, op *postgres.OperationStorage, kw *postgres.KeywordStorage,
) (*Bot, error) {
	bot := &Bot{
		state:     st,
		category:  cat,
		user:      usr,
		operation: op,
		keyword:   kw,
	}

	var err error

	bot.tele, err = tele.NewBot(tele.Settings{
		Token:     conf.AccessToken,
		Poller:    &tele.LongPoller{Timeout: time.Second},
		Verbose:   conf.Verbose,
		OnError:   bot.HandleError,
		ParseMode: tele.ModeHTML,
	})
	if err != nil {
		return nil, fmt.Errorf("telebot not created: %w", err)
	}

	if err = bot.setCommands(); err != nil {
		return nil, err
	}

	bot.tele.Use(middleware.Recover())
	bot.tele.Use(contextMiddleware(conf.Version))
	bot.tele.Use(bot.userContextMiddleware)

	bot.tele.Handle("/start", bot.start)

	bot.tele.Handle("/categories", bot.listCategories)
	bot.tele.Handle("/rename_category", bot.renameCategory)
	bot.tele.Handle("/add_category", bot.addCategory)
	bot.tele.Handle("/delete_keywords", bot.deleteKeywords)

	bot.tele.Handle("/set_language", bot.setLanguage)
	bot.tele.Handle("/set_currency", bot.setCurrency)

	bot.tele.Handle(tele.OnCallback, bot.handleCallback)
	bot.tele.Handle(tele.OnText, bot.handleText)

	return bot, nil
}

func (b *Bot) Start() {
	if b.tele != nil {
		b.tele.Start()
	}
}

func (b *Bot) Stop() {
	if b.tele != nil {
		b.tele.Stop()
	}
}

func (b *Bot) setCommands() error {
	err := b.tele.SetCommands([]tele.Command{
		{
			Text:        "categories",
			Description: "List categories",
		},

		{
			Text:        "add_category",
			Description: "Add new category",
		},
		{
			Text:        "rename_category",
			Description: "Rename category",
		},
		{
			Text:        "set_language",
			Description: "Change bot language",
		},
		{
			Text:        "set_currency",
			Description: "Change default currency",
		},
		{
			Text:        "delete_keywords",
			Description: "Delete operation keywords",
		},
	})
	if err != nil {
		return fmt.Errorf("commands not set: %w", err)
	}
	return nil
}

func btnCancel(kb *tele.ReplyMarkup, lang string) tele.Btn {
	return kb.Data(msg.Get(msg.BtnCancel, lang), botstate.StepCancel.String())
}

func (b *Bot) start(c tele.Context) error {
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

func (b *Bot) listCategories(c tele.Context) error {
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

	if _, err := sb.WriteString(fmt.Sprintf("%s\n\n", msg.Get(msg.ExpenseCatsTitle, usr.Language))); err != nil {
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

	if _, err := sb.WriteString(fmt.Sprintf("\n%s\n\n", msg.Get(msg.IncomeCatsTitle, usr.Language))); err != nil {
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

func (b *Bot) addCategory(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	kb := &tele.ReplyMarkup{}
	step := botstate.StepCatAddTypeSelection

	btnIncome := kb.Data(msg.Get(msg.BtnIncome, usr.Language), step.String(), domain.CatTypeIncome.String())
	btnExpenses := kb.Data(msg.Get(msg.BtnExpenses, usr.Language), step.String(), domain.CatTypeExpenses.String())

	kb.Inline(
		kb.Row(btnIncome),
		kb.Row(btnExpenses),
		kb.Row(btnCancel(kb, usr.Language)),
	)

	b.state.Add(usr.IDString(), botstate.State{Step: step})
	slog.InfoContext(stdContext(c), "added /add_category state", "step", step)

	return c.Send(msg.Get(msg.CatAddTypeSelection, usr.Language), kb)
}

func (b *Bot) renameCategory(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	kb := &tele.ReplyMarkup{}
	step := botstate.StepCatRenameTypeSelection

	btnIncome := kb.Data(msg.Get(msg.BtnIncome, usr.Language), step.String(), domain.CatTypeIncome.String())
	btnExpenses := kb.Data(msg.Get(msg.BtnExpenses, usr.Language), step.String(), domain.CatTypeExpenses.String())

	kb.Inline(
		kb.Row(btnIncome),
		kb.Row(btnExpenses),
		kb.Row(btnCancel(kb, usr.Language)),
	)

	b.state.Add(usr.IDString(), botstate.State{Step: step})

	return c.Send(msg.Get(msg.CatRenameTypeSelection, usr.Language), kb)
}

func (b *Bot) setLanguage(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	kb := &tele.ReplyMarkup{}
	step := botstate.StepLangSelection

	btnRus := kb.Data(msg.Get(msg.BtnRUS, usr.Language), step.String(), "ru")
	btnEng := kb.Data(msg.Get(msg.BtnENG, usr.Language), step.String(), "en")

	kb.Inline(
		kb.Row(btnRus),
		kb.Row(btnEng),
		kb.Row(btnCancel(kb, usr.Language)),
	)

	b.state.Add(usr.IDString(), botstate.State{Step: step})

	return c.Send(msg.Get(msg.LangSelection, usr.Language), kb)
}

func (b *Bot) setCurrency(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	kb := &tele.ReplyMarkup{}
	step := botstate.StepCurrSelection

	btnRUB := kb.Data(msg.Get(msg.BtnRUB, usr.Language), step.String(), "RUB")
	btnUSD := kb.Data(msg.Get(msg.BtnUSD, usr.Language), step.String(), "USD")
	btnEUR := kb.Data(msg.Get(msg.BtnEUR, usr.Language), step.String(), "EUR")

	kb.Inline(
		kb.Row(btnUSD),
		kb.Row(btnRUB),
		kb.Row(btnEUR),
		kb.Row(btnCancel(kb, usr.Language)),
	)

	b.state.Add(usr.IDString(), botstate.State{Step: step})

	return c.Send(msg.Get(msg.CurrSelection, usr.Language), kb)
}

func (b *Bot) handleText(c tele.Context) error {
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
	case botstate.StepCurrSelection:
		defer b.state.Remove(usr.IDString())

		// update usr currency
		usr.Currency = c.Text()

		if err := b.user.Update(ctx, usr); err != nil {
			if errors.Is(err, domain.ErrUnsupportedCurrency) {
				return c.Send(msg.Get(msg.InvalidCurr, usr.Language))
			}

			return fmt.Errorf("user not updated on text handle: %w", err)
		}

		return c.Send(msg.Getf(msg.CurrSaved, usr.Language, usr.Currency, usr.Currency))
	case botstate.StepCatRename:
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

		return c.Send(msg.Getf(msg.CatRenamed, usr.Language, cat.Name, newCatName))
	case botstate.StepCatAdd:
		defer b.state.Remove(usr.IDString())

		catType, ok := state.Data.(domain.CatType)
		if !ok {
			return errInvalidStateData
		}

		catName := c.Text()

		if err := b.category.SaveForUser(ctx, postgres.SaveCategoryParams{
			ID:        uuid.NewString(),
			Name:      catName,
			Type:      catType,
			Author:    usr.ID,
			CreatedAt: time.Now(),
		}); err != nil {
			return fmt.Errorf("category not created: %w", err)
		}

		slog.InfoContext(ctx, "new category added", "name", catName, "type", catType)

		return c.Send(msg.Getf(msg.CatAdded, usr.Language, catName))
	default:
		// handle operation save
		parts := strings.Split(c.Text(), " ")
		if len(parts) < 2 {
			return c.Send(msg.Get(msg.InvalidOperationFmt, usr.Language))
		}

		moneyStr := parts[0]

		// костыль
		if !strings.Contains(moneyStr, "-") && !strings.Contains(moneyStr, "+") {
			moneyStr = "-" + moneyStr
		}

		money, err := money.Parse(moneyStr)
		if err != nil {
			return c.Send(msg.Get(msg.InvalidOperationFmt, usr.Language))
		}

		if money == 0 {
			return c.Send(msg.Get(msg.InvalidOperationFmt, usr.Language))
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

		step := botstate.StepCatSelection

		kb, err := b.categoriesKeyboard(ctx, usr, step, catType, true)
		if err != nil {
			return err
		}

		b.state.Add(usr.IDString(), botstate.State{
			Step: step,
			Data: operation{
				name:      opName,
				money:     money,
				occuredAt: occuredAt,
			},
		})

		return c.Send(msg.Get(msg.CatSelection, usr.Language), kb)
	}
}

// categoriesKeyboard builds inline keyboard with categories.
func (b *Bot) categoriesKeyboard(ctx context.Context, usr domain.User, nextStep botstate.Step, ct domain.CatType, other bool) (*tele.ReplyMarkup, error) {
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

func (b *Bot) handleCallback(c tele.Context) error {
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

	switch botstate.Step(cb.unique) {
	case botstate.StepCatSelection:
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
	case botstate.StepCatRenameTypeSelection:
		kb, err := b.categoriesKeyboard(ctx, usr, botstate.StepCatRenameSelection, domain.CatType(cb.data), false)
		if err != nil {
			return fmt.Errorf("categories keyboard in callback: %w", err)
		}

		b.state.Add(usr.IDString(), botstate.State{Step: botstate.StepCatRenameSelection})

		return c.Edit(msg.Get(msg.CatRenameSelection, usr.Language), kb)
	case botstate.StepCatRenameSelection:
		cat, err := b.category.FindByID(ctx, cb.data)
		if err != nil {
			return fmt.Errorf("category not found on category rename: %w", err)
		}

		b.state.Add(usr.IDString(), botstate.State{
			Step: botstate.StepCatRename,
			Data: cat,
		})

		return c.Edit(msg.Getf(msg.CatRename, usr.Language, cat.Name))
	case botstate.StepCurrSelection:
		defer b.state.Remove(usr.IDString())

		// update usr currency
		usr.Currency = cb.data

		if err := b.user.Update(ctx, usr); err != nil {
			if errors.Is(err, domain.ErrUnsupportedCurrency) {
				return c.Send(msg.Get(msg.InvalidCurr, usr.Language))
			}

			return fmt.Errorf("currency not set in callback: %w", err)
		}

		return c.Edit(msg.Getf(msg.CurrSaved, usr.Language, usr.Currency, usr.Currency))
	case botstate.StepCatAddTypeSelection:
		b.state.Add(usr.IDString(), botstate.State{
			Step: botstate.StepCatAdd,
			Data: domain.CatType(cb.data),
		})
		slog.InfoContext(ctx, "added step on category add type selection", "step", botstate.StepCatAdd, "data", cb.data)
		return c.Edit(msg.Get(msg.CatAdd, usr.Language))
	case botstate.StepLangSelection:
		defer b.state.Remove(usr.IDString())

		// update usr languege
		usr.Language = cb.data

		if err := b.user.Update(ctx, usr); err != nil {
			return fmt.Errorf("language not set in callback: %w", err)
		}

		return c.Edit(msg.Getf(msg.LangSaved, usr.Language, iso6391.NativeName(usr.Language)))
	case botstate.StepCancel:
		b.state.Remove(usr.IDString())
		return c.Edit(msg.Get(msg.OperationCanceled, usr.Language))
	default:
		return fmt.Errorf("unsupported callback unique: %s", cb.unique)
	}
}

func (b *Bot) deleteKeywords(c tele.Context) error {
	usr, ok := userFromContext(c)
	if !ok {
		return errUserNotInContext
	}

	ctx := stdContext(c)

	if err := b.keyword.DeleteAll(ctx, usr.ID); err != nil {
		return fmt.Errorf("keywords not deleted: %w", err)
	}

	return c.Send(msg.Get(msg.KeywordsDeleted, usr.Language))
}

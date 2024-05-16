package msg

import "fmt"

type Type uint8

const (
	// general
	InternalError Type = iota
	OperationCanceled

	// steps
	CurrencySelection
	CurrencySaved

	LanguageSelection
	LanguageSaved

	CategorySelection
	ExpenseSaved
	IncomeSaved

	CategoryTypeSelection
	CategoryRenameSelection
	CategoryRename
	CategoryRenamed

	// logic errors
	InvalidCurrency
	InvalidOperationFormat

	// message titles
	ExpenseCategoriesTitle
	IncomeCategoriesTitle

	// buttons
	BtnCancel
	BtnRUB
	BtnEUR
	BtnUSD
	BtnRussian
	BtnEnglish
	BtnOther
	BtnIncome
	BtnExpenses
)

var messages = map[Type]map[string]string{
	// General
	InternalError: {
		"ru": "Произошла неизвестная ошибка, попробуйте позже",
		"en": "Internal server error, please try later",
	},
	OperationCanceled: {
		"ru": "Текущая операция отменена",
		"en": "Current operation is canceled",
	},

	// Steps
	CurrencySelection: {
		"ru": "Выбери из списка или отправь любую валюту в ISO-4217 формате (например UAH, KZT, GBP и другие)",
		"en": "Choose from list or send any other currency in ISO-4217 format (for example UAH, KZT, GBP etc)",
	},
	CurrencySaved: {
		"ru": "<b>%s</b> сохранена как валюта по умолчанию, для следующей команды без указания валюты я буду использовать <b>%s</b>",
		"en": "<b>%s</b> saved as your default currency, next time you send me a command without specifying currency I'll use <b>%s</b>",
	},
	CategorySelection: {
		"ru": "Выбери категорию расхода или дохода для учета",
		"en": "Choose category for the accounted expense or income",
	},
	LanguageSelection: {
		"ru": "Выбери язык с помощью которого я буду с тобой общаться",
		"en": "Select language which I'll be using for chatting with you",
	},
	LanguageSaved: {
		"ru": "Язык изменен на <b>%s</b>",
		"en": "Language was set to <b>%s</b>",
	},
	ExpenseSaved: {
		"ru": "Потрачено <b>%s %s</b> в категории %s\n\n<i>%s</i>",
		"en": "Spent <b>%s %s</b> in %s category\n\n<i>%s</i>",
	},
	IncomeSaved: {
		"ru": "Заработано <b>%s %s</b> в категории %s\n\n<i>%s</i>",
		"en": "Earned <b>%s %s</b> in %s category\n\n<i>%s</i>",
	},
	CategoryTypeSelection: {
		"ru": "Категорию расходов или доходов хочешь переименовать?",
		"en": "Category of expenses or income would like to rename?",
	},
	CategoryRenameSelection: {
		"ru": "Какую категорию хочешь переименовать?",
		"en": "Which category you want to rename?",
	},
	CategoryRename: {
		"ru": "Как теперь будет называться категория <b>%s</b>?",
		"en": "What will <b>%s</b> category be called now?",
	},
	CategoryRenamed: {
		"ru": "Категория <b>%s</b> переименована в <b>%s</b>",
		"en": "Category <b>%s</b> renamed to <b>%s</b>",
	},

	// Logic errors
	InvalidCurrency: {
		"ru": "Некорректный формат валюты, отправь валюту в ISO-4217 формате",
		"en": "Invalid currency format, provide currency code in ISO-4217 format",
	},

	// Message titles
	ExpenseCategoriesTitle: {
		"ru": "➖ Категории расходов",
		"en": "➖ Expense categories",
	},
	IncomeCategoriesTitle: {
		"ru": "➕ Категории доходов",
		"en": "➕ Income categories",
	},
	InvalidOperationFormat: {
		"ru": "Некорректный формат операции",
		"en": "Invalid operation format",
	},

	// buttons
	BtnRUB: {
		"ru": "🇷🇺 Рубли",
		"en": "🇷🇺 Rubles",
	},
	BtnUSD: {
		"ru": "🇺🇸 Американские доллары",
		"en": "🇺🇸 Dollars",
	},
	BtnEUR: {
		"ru": "🇪🇺 Евро",
		"en": "🇪🇺 Euros",
	},
	BtnCancel: {
		"ru": "Отмена",
		"en": "Cancel",
	},
	BtnRussian: {
		"ru": "🇷🇺 Русский",
		"en": "🇷🇺 Russian",
	},
	BtnEnglish: {
		"ru": "🇺🇸 Английский",
		"en": "🇺🇸 English",
	},
	BtnOther: {
		"ru": "🤷 Другое",
		"en": "🤷 Other",
	},
	BtnIncome: {
		"ru": "📈 Доходы",
		"en": "📈 Income",
	},
	BtnExpenses: {
		"ru": "📉 Расходы",
		"en": "📉 Expenses",
	},
}

func Get(t Type, lang string) string {
	msg, ok := messages[t][lang]
	if !ok {
		return messages[InternalError]["en"]
	}

	return msg
}

func Getf(t Type, lang string, args ...any) string {
	msg := Get(t, lang)
	return fmt.Sprintf(msg, args...)
}

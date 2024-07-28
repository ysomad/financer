package msg

import "fmt"

type ID uint8

const (
	InternalError ID = iota
	OperationCanceled

	// Currency set
	CurrSelection
	CurrSaved

	// Language set
	LangSelection
	LangSaved

	// Operation create
	CatSelection
	ExpenseSaved
	IncomeSaved

	// Category rename
	CatRenameTypeSelection
	CatRenameSelection
	CatRename
	CatRenamed

	// Category add
	CatAddTypeSelection
	CatAdd
	CatAdded

	KeywordsDeleted

	// logic errors
	InvalidCurr
	InvalidOperationFmt

	// message titles
	ExpenseCatsTitle
	IncomeCatsTitle

	// buttons
	BtnCancel
	BtnRUB
	BtnEUR
	BtnUSD

	BtnRUS
	BtnENG

	BtnOther
	BtnIncome
	BtnExpenses
)

type Message struct {
	RU string
	EN string
}

var messages = map[ID]Message{
	// General
	InternalError: {
		RU: "Произошла неизвестная ошибка, попробуйте позже",
		EN: "Internal server error, please try later",
	},
	OperationCanceled: {
		RU: "Текущая операция отменена",
		EN: "Current operation is canceled",
	},

	// Steps
	CurrSelection: {
		RU: "Выбери из списка или отправь любую валюту в ISO-4217 формате (например UAH, KZT, GBP и другие)",
		EN: "Choose from list or send any other currency in ISO-4217 format (for example UAH, KZT, GBP etc)",
	},
	CurrSaved: {
		RU: "<b>%s</b> сохранена как валюта по умолчанию, для следующей команды без указания валюты я буду использовать <b>%s</b>",
		EN: "<b>%s</b> saved as your default currency, next time you send me a command without specifying currency I'll use <b>%s</b>",
	},
	CatSelection: {
		RU: "Выбери категорию расхода или дохода для учета",
		EN: "Choose category for the accounted expense or income",
	},
	LangSelection: {
		RU: "Выбери язык с помощью которого я буду с тобой общаться",
		EN: "Select language which I'll be using for chatting with you",
	},
	LangSaved: {
		RU: "Язык изменен на <b>%s</b>",
		EN: "Language was set to <b>%s</b>",
	},
	ExpenseSaved: {
		RU: "Потрачено <b>%s %s</b> в категории %s\n\n<i>%s</i>",
		EN: "Spent <b>%s %s</b> in %s category\n\n<i>%s</i>",
	},
	IncomeSaved: {
		RU: "Заработано <b>%s %s</b> в категории %s\n\n<i>%s</i>",
		EN: "Earned <b>%s %s</b> in %s category\n\n<i>%s</i>",
	},
	CatRenameTypeSelection: {
		RU: "Категорию расходов или доходов хочешь переименовать?",
		EN: "Category of expenses or income would like to rename?",
	},
	CatRenameSelection: {
		RU: "Какую категорию хочешь переименовать?",
		EN: "Which category you want to rename?",
	},
	CatRename: {
		RU: "Как теперь будет называться категория <b>%s</b>?",
		EN: "What will <b>%s</b> category be called now?",
	},
	CatRenamed: {
		RU: "Категория <b>%s</b> переименована в <b>%s</b>",
		EN: "Category <b>%s</b> renamed to <b>%s</b>",
	},
	CatAddTypeSelection: {
		RU: "Категорию расходов или доходов хочешь добавить?",
		EN: "Category of expenses or income would like to add?",
	},
	CatAdd: {
		RU: "Как будет называться новая категория?",
		EN: "What will new category be called?",
	},
	CatAdded: {
		RU: "Категория <b>%s</b> успешно создана",
		EN: "Category <b>%s</b> successfully created",
	},
	KeywordsDeleted: {
		RU: "Все ключевые слова операций удалены",
		EN: "All operation keywords deleted",
	},

	// Logic errors
	InvalidCurr: {
		RU: "Некорректный формат валюты, отправь валюту в ISO-4217 формате",
		EN: "Invalid currency format, provide currency code in ISO-4217 format",
	},

	// Message titles
	ExpenseCatsTitle: {
		RU: "➖ Категории расходов",
		EN: "➖ Expense categories",
	},
	IncomeCatsTitle: {
		RU: "➕ Категории доходов",
		EN: "➕ Income categories",
	},
	InvalidOperationFmt: {
		RU: "Некорректный формат операции",
		EN: "Invalid operation format",
	},

	// Buttons
	BtnRUB: {
		RU: "🇷🇺 Рубли",
		EN: "🇷🇺 Rubles",
	},
	BtnUSD: {
		RU: "🇺🇸 Американские доллары",
		EN: "🇺🇸 Dollars",
	},
	BtnEUR: {
		RU: "🇪🇺 Евро",
		EN: "🇪🇺 Euros",
	},
	BtnCancel: {
		RU: "Отмена",
		EN: "Cancel",
	},
	BtnRUS: {
		RU: "🇷🇺 Русский",
		EN: "🇷🇺 Russian",
	},
	BtnENG: {
		RU: "🇺🇸 Английский",
		EN: "🇺🇸 English",
	},
	BtnOther: {
		RU: "🤷 Другое",
		EN: "🤷 Other",
	},
	BtnIncome: {
		RU: "📈 Доходы",
		EN: "📈 Income",
	},
	BtnExpenses: {
		RU: "📉 Расходы",
		EN: "📉 Expenses",
	},
}

func Get(id ID, lang string) string {
	msg, ok := messages[id]
	if !ok {
		return messages[InternalError].EN
	}
	switch lang {
	case "en":
		return msg.EN
	case "ru":
		return msg.RU
	default:
		return msg.EN
	}
}

func Getf(id ID, lang string, args ...any) string {
	return fmt.Sprintf(Get(id, lang), args...)
}

package domain

type State struct {
	Step Step
	Data any
}

type Step string

const (
	StepUnknown Step = "unknown"
	StepCancel  Step = "cancel"

	StepCurrencySelection Step = "currency_selection"
	StepCategorySelection Step = "category_selection"
	StepLanguageSelection Step = "language_selection"

	StepCategoryTypeSelection   Step = "category_type_selection"
	StepCategoryRenameSelection Step = "category_rename_selection"
	StepCategoryRename          Step = "category_rename"
)

func (s Step) String() string {
	return string(s)
}

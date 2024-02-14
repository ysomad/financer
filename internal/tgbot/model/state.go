package model

type State struct {
	Step Step
	Data []byte
}

type Step string

// TODO: move to its own package
const (
	StepUnknown           Step = "unknown"
	StepCancel            Step = "cancel"
	StepCurrencySelection Step = "currency_selection"
	StepCategorySelection Step = "category_selection"
)

func (s Step) String() string {
	return string(s)
}

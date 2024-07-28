package state

type State struct {
	Step Step
	Data any
}

type Step string

const (
	StepUnknown Step = "unknown"
	StepCancel  Step = "cancel"

	// Operation create
	StepCurrSelection Step = "currency_selection"

	StepCatSelection  Step = "category_selection"
	StepLangSelection Step = "language_selection"

	// Category rename
	StepCatRenameTypeSelection Step = "category_rename_type_selection"
	StepCatRenameSelection     Step = "category_rename_selection"
	StepCatRename              Step = "category_rename"

	// Category add
	StepCatAddTypeSelection Step = "category_add_type_selection"
	StepCatAdd              Step = "category_step_add"
)

func (s Step) String() string {
	return string(s)
}

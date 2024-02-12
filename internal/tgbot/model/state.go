package model

type State string

// TODO: move to its own package
const (
	StateUnknown           State = "unknown"
	StateCurrencySelection State = "currency_selection"
)

package domain

import "github.com/google/uuid"

type CatType string

const (
	CatTypeUnspecified = ""
	CatTypeExpense     = "EXPENSE"
	CatTypeIncome      = "INCOME"
)

var OtherCategoryID = uuid.Nil.String()

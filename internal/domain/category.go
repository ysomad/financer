package domain

import "github.com/google/uuid"

type CatType string

const (
	CatTypeUnspecified CatType = ""
	CatTypeExpense     CatType = "EXPENSE"
	CatTypeIncome      CatType = "INCOME"
)

var OtherCategoryID = uuid.Nil.String()

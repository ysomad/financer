package domain

import "github.com/google/uuid"

type CatType string

const (
	CatTypeUnspecified CatType = ""
	CatTypeExpenses    CatType = "EXPENSES"
	CatTypeIncome      CatType = "INCOME"
	CatTypeOther       CatType = "OTHER"
)

func (t CatType) String() string {
	return string(t)
}

var OtherCategoryID = uuid.Nil.String()

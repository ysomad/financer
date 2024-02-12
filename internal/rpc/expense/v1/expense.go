package v1

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/ladydascalie/currency"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"

	"github.com/ysomad/financer/internal/auth"
	pb "github.com/ysomad/financer/internal/gen/proto/expense/v1"
	pbconnect "github.com/ysomad/financer/internal/gen/proto/expense/v1/expensev1connect"
	"github.com/ysomad/financer/internal/guid"
	"github.com/ysomad/financer/internal/postgres"
)

var _ pbconnect.ExpenseServiceHandler = &ExpenseServer{}

type ExpenseServer struct {
	pbconnect.UnimplementedExpenseServiceHandler // TODO: remove after implement
	expense                                      postgres.ExpenseStorage
}

func NewExpenseServer(s postgres.ExpenseStorage) *ExpenseServer {
	return &ExpenseServer{expense: s}
}

var errInvalidCurrencyCode = errors.New("currency code must be valid ISO-4217 code")

func (s *ExpenseServer) FindExpense(ctx context.Context, r *connect.Request[pb.FindExpenseRequest]) (*connect.Response[pb.Expense], error) {
	expense, err := s.expense.Find(ctx, auth.IdentityID(ctx), r.Msg.ExpenseName)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, postgres.ErrNotFound)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.Expense{
		Id: expense.ID,
		Money: &money.Money{
			CurrencyCode: expense.Currency,
			Units:        expense.MoneyUnits,
			Nanos:        expense.MoneyNanos,
		},
		Name:     expense.Name,
		Category: expense.Category,
		Date: &date.Date{
			Year:  int32(expense.Date.Year()),
			Month: int32(expense.Date.Month()),
			Day:   int32(expense.Date.Day()),
		},
	}), nil
}

func (s *ExpenseServer) DeclareExpense(ctx context.Context, r *connect.Request[pb.DeclareExpenseRequest]) (*connect.Response[pb.Expense], error) {
	// TODO: move to protovalidate
	// not empty and valid
	if !currency.Valid(r.Msg.Money.CurrencyCode) && r.Msg.Money.CurrencyCode != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errInvalidCurrencyCode)
	}

	p := postgres.SaveExpenseParams{
		ID:         guid.New("expense"),
		IdentityID: auth.IdentityID(ctx),
		Category:   r.Msg.Category,
		Name:       r.Msg.Name,
		Currency:   r.Msg.Money.CurrencyCode,
		MoneyUnits: r.Msg.Money.Units,
		MoneyNanos: r.Msg.Money.Nanos,
		Date:       time.Date(int(r.Msg.Date.Year), time.Month(r.Msg.Date.Month), int(r.Msg.Date.Day), 0, 0, 0, 0, time.UTC),
		CreatedAt:  time.Now(),
	}

	if err := s.expense.Save(ctx, p); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&pb.Expense{
		Id:       p.ID,
		Money:    r.Msg.Money,
		Name:     r.Msg.Name,
		Category: r.Msg.Category,
		Date:     r.Msg.Date,
	}), nil
}

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
	identity                                     postgres.IdentityStorage
}

func NewExpenseServer(s postgres.ExpenseStorage, i postgres.IdentityStorage) *ExpenseServer {
	return &ExpenseServer{
		expense:  s,
		identity: i,
	}
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
	var (
		curr = r.Msg.Money.CurrencyCode
		date time.Time
		err  error
	)

	// TODO: move to protovalidate
	// not empty and valid
	if !currency.Valid(curr) && curr != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errInvalidCurrencyCode)
	}

	identityID := auth.IdentityID(ctx)

	// use default currency if its not provided
	if curr == "" {
		curr, err = s.identity.GetCurrency(ctx, identityID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// if date of expense is not provided
	if r.Msg.Date.Day == 0 || r.Msg.Date.Month == 0 || r.Msg.Date.Year == 0 {
		date = time.Now()
	} else {
		date = time.Date(int(r.Msg.Date.Year), time.Month(r.Msg.Date.Month), int(r.Msg.Date.Day), 0, 0, 0, 0, time.UTC)
	}

	p := postgres.SaveExpenseParams{
		ID:         guid.New("expense"),
		IdentityID: identityID,
		Category:   r.Msg.Category,
		Name:       r.Msg.Name,
		Currency:   curr,
		MoneyUnits: r.Msg.Money.Units,
		MoneyNanos: r.Msg.Money.Nanos,
		Date:       date,
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

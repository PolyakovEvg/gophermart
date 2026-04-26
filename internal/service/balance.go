package service

import (
	"context"

	"go-musthave-diploma-tpl/internal/repository/postgres"

	"github.com/shopspring/decimal"
)

type BalanceServicer interface {
	GetBalance(ctx context.Context, userID string) (current, withdrawn decimal.Decimal, err error)
	Withdraw(ctx context.Context, userID, orderNumber string, sum decimal.Decimal) error
	ListWithdrawals(ctx context.Context, userID string) ([]postgres.Withdrawal, error)
}
type BalanceService struct {
	withdrawRepo *postgres.WithdrawalRepository
}

func NewBalanceService(withdrawRepo *postgres.WithdrawalRepository) *BalanceService {
	return &BalanceService{
		withdrawRepo: withdrawRepo,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID string,
) (current, withdrawn decimal.Decimal, err error) {

	accrued, withdrawn, err := s.withdrawRepo.GetTotals(ctx, userID)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	current = accrued.Sub(withdrawn)
	return current, withdrawn, nil
}

func (s *BalanceService) Withdraw(ctx context.Context, userID, orderNumber string, sum decimal.Decimal,
) error {
	if sum.LessThanOrEqual(decimal.Zero) {
		return postgres.ErrInvalidOrder
	}

	current, _, err := s.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if current.LessThan(sum) {
		return postgres.ErrNotEnoughFunds
	}

	return s.withdrawRepo.Create(ctx, userID, orderNumber, sum)
}

func (s *BalanceService) ListWithdrawals(ctx context.Context, userID string,
) ([]postgres.Withdrawal, error) {
	return s.withdrawRepo.ListByUser(ctx, userID)
}

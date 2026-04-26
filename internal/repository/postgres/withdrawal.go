package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrNotEnoughFunds = errors.New("not enough funds")
)

type Withdrawal struct {
	ID          string
	UserID      string
	OrderNumber string
	Sum         decimal.Decimal
	ProcessedAt time.Time
}

type WithdrawalRepository struct {
	db *sql.DB
}

func NewWithdrawalRepository(db *sql.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

func (r *WithdrawalRepository) Create(
	ctx context.Context,
	userID, orderNumber string,
	sum decimal.Decimal,
) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, userID, orderNumber, sum)
	if err != nil {
		if strings.Contains(err.Error(), "withdrawals_user_id_order_number_key") {
			return ErrInvalidOrder
		}
		return err
	}
	return nil
}

func (r *WithdrawalRepository) ListByUser(
	ctx context.Context,
	userID string,
) ([]Withdrawal, error) {
	query := `
		SELECT id, user_id, order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Withdrawal
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.OrderNumber,
			&w.Sum,
			&w.ProcessedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *WithdrawalRepository) GetTotals(
	ctx context.Context,
	userID string,
) (accrued, withdrawn decimal.Decimal, err error) {
	query := `
		SELECT
			COALESCE(SUM(o.accrual), 0),
			COALESCE((
				SELECT SUM(w.sum)
				FROM withdrawals w
				WHERE w.user_id = $1
			), 0)
		FROM orders o
		WHERE o.user_id = $1
		  AND o.status = 'PROCESSED'
	`

	err = r.db.QueryRowContext(ctx, query, userID).Scan(&accrued, &withdrawn)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	return accrued, withdrawn, nil
}

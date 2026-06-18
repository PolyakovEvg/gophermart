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
	ErrOrderExists         = errors.New("order already exists")
	ErrOrderUploadedByUser = errors.New("order already uploaded by this user")
	ErrInvalidOrder        = errors.New("invalid order number")
)

type OrderRepository struct {
	db *sql.DB
}

type Order struct {
	ID         string              `db:"id"`
	Number     string              `db:"number"`
	UserID     string              `db:"user_id"`
	Status     string              `db:"status"`
	Accrual    decimal.NullDecimal `db:"accrual"`
	UploadedAt time.Time           `db:"uploaded_at"`
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, userID, number string) error {
	query := `INSERT INTO orders (user_id, number) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, query, userID, number)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return ErrOrderExists
		}
		return err
	}
	return nil
}

func (r *OrderRepository) GetOrderByNumber(ctx context.Context, number string) (*Order, error) {
	query := `
        SELECT id, number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE number = $1
    `
	var o Order
	err := r.db.QueryRowContext(ctx, query, number).Scan(
		&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) GetOrderByUser(ctx context.Context, userID string) ([]Order, error) {
	query := `
        SELECT id, number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE user_id = $1
        ORDER BY uploaded_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) GetOrdersForProcessing(ctx context.Context, limit int) ([]Order, error) {
	query := `
        SELECT id, number, user_id, status, accrual, uploaded_at
        FROM orders
        WHERE status IN ('NEW', 'PROCESSING')
        ORDER BY uploaded_at
        LIMIT $1
    `
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Number, &o.UserID, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string, accrual *decimal.Decimal) error {
	var dbAccrual decimal.NullDecimal
	if accrual != nil {
		dbAccrual = decimal.NullDecimal{
			Decimal: *accrual,
			Valid:   true,
		}
	} else {
		dbAccrual = decimal.NullDecimal{Valid: false}
	}

	query := `UPDATE orders SET status = $2, accrual = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, orderID, status, dbAccrual)
	return err
}

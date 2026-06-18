package service_test

import (
	"context"
	"errors"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type mockOrdersRepo struct {
	CreateFunc      func(ctx context.Context, userID, number string) error
	GetByNumberFunc func(ctx context.Context, number string) (*postgres.Order, error)
	GetByUserFunc   func(ctx context.Context, userID string) ([]postgres.Order, error)
}

func (m *mockOrdersRepo) CreateOrder(ctx context.Context, userID, number string) error {
	return m.CreateFunc(ctx, userID, number)
}

func (m *mockOrdersRepo) GetOrderByNumber(ctx context.Context, number string) (*postgres.Order, error) {
	return m.GetByNumberFunc(ctx, number)
}

func (m *mockOrdersRepo) GetOrderByUser(ctx context.Context, userID string) ([]postgres.Order, error) {
	return m.GetByUserFunc(ctx, userID)
}

func TestOrdersService_UploadOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByNumberFunc: func(ctx context.Context, number string) (*postgres.Order, error) {
				return nil, nil
			},
			CreateFunc: func(ctx context.Context, userID, number string) error {
				return nil
			},
		}
		svc := service.NewOrdersService(mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.NoError(t, err)
	})

	t.Run("empty_number", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{}
		svc := service.NewOrdersService(mockRepo)

		err := svc.UploadOrder(ctx, "user1", "")
		require.Error(t, err)
		require.Equal(t, "order number required", err.Error())
	})

	t.Run("invalid_luhn", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{}
		svc := service.NewOrdersService(mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678901")
		require.ErrorIs(t, err, postgres.ErrInvalidOrder)
	})

	t.Run("already_uploaded_by_same_user", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByNumberFunc: func(ctx context.Context, number string) (*postgres.Order, error) {
				return &postgres.Order{
					ID:     "1",
					Number: "12345678903",
					UserID: "user1",
					Status: "NEW",
				}, nil
			},
		}
		svc := service.NewOrdersService(mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.ErrorIs(t, err, service.ErrOrderAlreadyUploaded)
	})

	t.Run("already_uploaded_by_other_user", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByNumberFunc: func(ctx context.Context, number string) (*postgres.Order, error) {
				return &postgres.Order{
					ID:     "1",
					Number: "12345678903",
					UserID: "user2",
					Status: "NEW",
				}, nil
			},
		}
		svc := service.NewOrdersService(mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.ErrorIs(t, err, postgres.ErrOrderExists)
	})

	t.Run("create_error", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByNumberFunc: func(ctx context.Context, number string) (*postgres.Order, error) {
				return nil, nil
			},
			CreateFunc: func(ctx context.Context, userID, number string) error {
				return errors.New("db error")
			},
		}
		svc := service.NewOrdersService(mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.Error(t, err)
		require.Equal(t, "db error", err.Error())
	})

	t.Run("race_condition_same_user", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByNumberFunc: func(ctx context.Context, number string) (*postgres.Order, error) {
				return nil, nil
			},
			CreateFunc: func(ctx context.Context, userID, number string) error {
				return postgres.ErrOrderExists
			},
		}
		svc := service.NewOrdersService(mockRepo)

		err := svc.UploadOrder(ctx, "user1", "12345678903")
		require.ErrorIs(t, err, postgres.ErrOrderExists)
	})
}

func TestOrdersService_ListOrders(t *testing.T) {
	ctx := context.Background()

	t.Run("orders_exist", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByUserFunc: func(ctx context.Context, userID string) ([]postgres.Order, error) {
				return []postgres.Order{
					{
						ID:         "1",
						Number:     "12345678903",
						UserID:     userID,
						Status:     "NEW",
						UploadedAt: time.Now(),
					},
				}, nil
			},
		}
		svc := service.NewOrdersService(mockRepo)

		orders, err := svc.ListOrders(ctx, "user1")
		require.NoError(t, err)
		require.Len(t, orders, 1)
		require.Equal(t, "12345678903", orders[0].Number)
	})

	t.Run("orders_with_accrual", func(t *testing.T) {
		accrual := decimal.NewFromFloat(42.50)
		mockRepo := &mockOrdersRepo{
			GetByUserFunc: func(ctx context.Context, userID string) ([]postgres.Order, error) {
				return []postgres.Order{
					{
						ID:         "1",
						Number:     "12345678903",
						UserID:     userID,
						Status:     "PROCESSED",
						Accrual:    decimal.NullDecimal{Decimal: accrual, Valid: true},
						UploadedAt: time.Now(),
					},
				}, nil
			},
		}
		svc := service.NewOrdersService(mockRepo)

		orders, err := svc.ListOrders(ctx, "user1")
		require.NoError(t, err)
		require.Len(t, orders, 1)
		require.True(t, orders[0].Accrual.Valid)
		require.True(t, orders[0].Accrual.Decimal.Equal(decimal.NewFromFloat(42.5)))
	})

	t.Run("no_orders", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByUserFunc: func(ctx context.Context, userID string) ([]postgres.Order, error) {
				return []postgres.Order{}, nil
			},
		}
		svc := service.NewOrdersService(mockRepo)

		orders, err := svc.ListOrders(ctx, "user1")
		require.NoError(t, err)
		require.Len(t, orders, 0)
	})

	t.Run("error_from_repo", func(t *testing.T) {
		mockRepo := &mockOrdersRepo{
			GetByUserFunc: func(ctx context.Context, userID string) ([]postgres.Order, error) {
				return nil, errors.New("db error")
			},
		}
		svc := service.NewOrdersService(mockRepo)

		orders, err := svc.ListOrders(ctx, "user1")
		require.Error(t, err)
		require.Nil(t, orders)
	})
}
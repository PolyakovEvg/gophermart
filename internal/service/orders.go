package service

import (
	"context"
	"errors"

	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/pkg/luhn"

	"go.uber.org/zap"
)

var ErrOrderAlreadyUploaded = errors.New("order already uploaded by user")

type OrdersServicer interface {
	CreateOrder(ctx context.Context, userID, number string, logger *zap.Logger) error
	GetOrderByNumber(ctx context.Context, number string, logger *zap.Logger) (*postgres.Order, error)
	GetOrderByUser(ctx context.Context, userID string, logger *zap.Logger) ([]postgres.Order, error)
}

type OrdersService struct {
	orderRepo OrdersServicer
	logger    *zap.Logger
}

func NewOrdersService(logger *zap.Logger, orderRepo OrdersServicer) *OrdersService {
	return &OrdersService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (s *OrdersService) UploadOrder(ctx context.Context, userID, number string) error {
	if number == "" {
		return errors.New("order number required")
	}
	if !luhn.ValidString(number) {
		return postgres.ErrInvalidOrder
	}

	existing, err := s.orderRepo.GetOrderByNumber(ctx, number, s.logger)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.UserID == userID {
			return ErrOrderAlreadyUploaded
		}
		return postgres.ErrOrderExists
	}

	err = s.orderRepo.CreateOrder(ctx, userID, number, s.logger)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderExists) {
			existing, checkErr := s.orderRepo.GetOrderByNumber(ctx, number, s.logger)
			if checkErr == nil && existing != nil {
				if existing.UserID == userID {
					return ErrOrderAlreadyUploaded
				}
				return postgres.ErrOrderExists
			}
		}
		return err
	}

	return nil
}

func (s *OrdersService) ListOrders(ctx context.Context, userID string) ([]postgres.Order, error) {
	return s.orderRepo.GetOrderByUser(ctx, userID, s.logger)
}

package service

import (
	"context"
	"errors"

	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/pkg/luhn"
)

var ErrOrderAlreadyUploaded = errors.New("order already uploaded by user")

type OrdersServicer interface {
	CreateOrder(ctx context.Context, userID, number string) error
	GetOrderByNumber(ctx context.Context, number string) (*postgres.Order, error)
	GetOrderByUser(ctx context.Context, userID string) ([]postgres.Order, error)
}

type OrdersService struct {
	orderRepo OrdersServicer
}

func NewOrdersService(orderRepo OrdersServicer) *OrdersService {
	return &OrdersService{
		orderRepo: orderRepo,
	}
}

func (s *OrdersService) UploadOrder(ctx context.Context, userID, number string) error {
	if number == "" {
		return errors.New("order number required")
	}
	if !luhn.ValidString(number) {
		return postgres.ErrInvalidOrder
	}

	existing, err := s.orderRepo.GetOrderByNumber(ctx, number)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.UserID == userID {
			return ErrOrderAlreadyUploaded
		}
		return postgres.ErrOrderExists
	}

	err = s.orderRepo.CreateOrder(ctx, userID, number)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderExists) {
			existing, checkErr := s.orderRepo.GetOrderByNumber(ctx, number)
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
	return s.orderRepo.GetOrderByUser(ctx, userID)
}

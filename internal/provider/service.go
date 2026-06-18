package provider

import (
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"
)

func NewAuthService(repo *postgres.UserRepository, cfg *config.Config) *service.AuthService {
	return service.NewAuthService(repo, cfg.AuthSecret)
}

func NewOrdersService(orderRepo *postgres.OrderRepository) *service.OrdersService {
	return service.NewOrdersService(orderRepo)
}

func NewBalanceService(repo *postgres.WithdrawalRepository) *service.BalanceService {
	return service.NewBalanceService(repo)
}

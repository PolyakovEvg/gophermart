package provider

import (
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"

	"go.uber.org/zap"
)

func NewAuthService(repo *postgres.UserRepository, cfg *config.Config, logger *zap.Logger) *service.AuthService {
	return service.NewAuthService(repo, cfg.AuthSecret, logger)
}

func NewOrdersService(logger *zap.Logger, orderRepo *postgres.OrderRepository) *service.OrdersService {
	return service.NewOrdersService(logger, orderRepo)
}

func NewBalanceService(repo *postgres.WithdrawalRepository, logger *zap.Logger) *service.BalanceService {
	return service.NewBalanceService(repo, logger)
}

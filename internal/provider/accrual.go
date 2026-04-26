package provider

import (
	"go-musthave-diploma-tpl/internal/accrual"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"

	"go.uber.org/zap"
)

func NewAccrualClient(cfg *config.Config) *accrual.Client {
	return service.NewAccrualClient(cfg)
}

func NewAccrualWorker(orderRepo *postgres.OrderRepository, client *accrual.Client, logger *zap.Logger) *service.AccrualWorker {
	return service.NewAccrualWorker(orderRepo, client, logger)
}

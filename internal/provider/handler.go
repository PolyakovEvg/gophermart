package provider

import (
	"go-musthave-diploma-tpl/internal/handler"
	"go-musthave-diploma-tpl/internal/service"

	"go.uber.org/zap"
)

func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *handler.AuthHandler {
	return handler.NewAuthHandler(authService, logger)
}

func NewOrdersHandler(ordersService *service.OrdersService, logger *zap.Logger) *handler.OrdersHandler {
	return handler.NewOrdersHandler(ordersService, logger)
}

func NewBalanceHandler(s *service.BalanceService, logger *zap.Logger) *handler.BalanceHandler {
	return handler.NewBalanceHandler(s, logger)
}

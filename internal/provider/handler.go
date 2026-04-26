package provider

import (
	"go-musthave-diploma-tpl/internal/handler"
	"go-musthave-diploma-tpl/internal/service"

	"go.uber.org/zap"
)

func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *handler.AuthHandler {
	return handler.NewAuthHandler(authService, logger)
}

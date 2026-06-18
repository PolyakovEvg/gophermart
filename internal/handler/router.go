package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/config"
	customMiddleware "go-musthave-diploma-tpl/internal/middleware"
)

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	authHandler *AuthHandler,
	ordersHandler *OrdersHandler,
	balanceHandler *BalanceHandler,
) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.Logger(logger))

	r.Post("/api/user/register", authHandler.Register)
	r.Post("/api/user/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(cfg.AuthSecret, logger))
		r.Post("/api/user/orders", ordersHandler.UploadOrder)
		r.Get("/api/user/orders", ordersHandler.ListOrders)
		r.Get("/api/user/balance", balanceHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
		r.Get("/api/user/withdrawals", balanceHandler.ListWithdrawals)
	})

	return r
}

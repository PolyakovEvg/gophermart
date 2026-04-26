package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/handler"
	"go-musthave-diploma-tpl/internal/provider"
	"go-musthave-diploma-tpl/internal/repository/postgres"
	"go-musthave-diploma-tpl/internal/service"
)

type App struct {
	*fx.App
}

func New() *App {
	app := fx.New(
		fx.Provide(
			config.InitConfig,
			provider.NewLogger,
			provider.NewStorage,
			provider.NewUserRepository,
			provider.NewAuthService,
			provider.NewAuthHandler,
			handler.NewRouter,
		),
		fx.Invoke(
			startServer,
			runMigrations,
		),
	)

	return &App{App: app}
}

func (a *App) Run() error {
	a.App.Run()
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	return a.App.Stop(ctx)
}

func startServer(lc fx.Lifecycle, cfg *config.Config, router chi.Router, logger *zap.Logger) {
	srv := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting HTTP server", zap.String("addr", cfg.RunAddress))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("HTTP server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}

func runMigrations(cfg *config.Config, logger *zap.Logger) error {
	if err := postgres.RunMigrations(cfg.DatabaseURI, logger); err != nil {
		logger.Error("migrations failed", zap.Error(err))
		return err
	}
	logger.Info("migrations completed successfully")
	return nil
}

func StartAccrualWorker(lc fx.Lifecycle, worker *service.AccrualWorker, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting accrual worker")

			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						logger.Info("accrual worker stopped", zap.Error(ctx.Err()))
						return
					case <-ticker.C:
						worker.Process(ctx)
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping accrual worker")
			return nil
		},
	})
}

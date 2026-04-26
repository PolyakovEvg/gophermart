package provider

import (
	"context"
	"fmt"

	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/repository/postgres"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewStorage(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) (*postgres.DBStorage, error) {
	dbStore, err := postgres.NewDBStorage(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return dbStore.Close()
		},
	})
	return dbStore, nil
}

package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type DBStorage struct {
	DB     *sql.DB
	Logger *zap.Logger
}

func NewDBStorage(dsn string, logger *zap.Logger) (*DBStorage, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("failed to open database", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("failed to ping database", zap.Error(err))
		return nil, err
	}

	logger.Info("successfully connected to database", zap.String("dns", dsn))

	return &DBStorage{
		DB:     db,
		Logger: logger,
	}, nil
}

func (s *DBStorage) PingContext(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *DBStorage) Close() error {
	return s.DB.Close()
}

package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	DB *sql.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DBStorage{DB: db}, nil
}

func (s *DBStorage) PingContext(ctx context.Context) error {
	return s.DB.PingContext(ctx)
}

func (s *DBStorage) Close() error {
	return s.DB.Close()
}

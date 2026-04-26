package postgres

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func RunMigrations(uri string, logger *zap.Logger) error {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		logger.Error("cannot open DB", zap.Error(err))
		return err
	}
	defer db.Close()

	path, err := filepath.Abs("migrations")
	if err != nil {
		logger.Error("failed to get absolute path", zap.Error(err))
		return err
	}

	migrationsURL := "file://" + path
	logger.Info("Running migrations", zap.String("path", migrationsURL))

	m, err := migrate.New(migrationsURL, uri)
	if err != nil {
		logger.Error("cannot create migration", zap.Error(err))
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Info("cannot run migration", zap.Error(err))
		return err
	}
	logger.Info("migrations successfully migrated")

	return nil
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

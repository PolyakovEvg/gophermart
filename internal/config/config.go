package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AuthSecret           string `env:"AUTH_SECRET"`
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"--a %s --d %s --r %s",
		c.RunAddress,
		c.DatabaseURI,
		c.AccrualSystemAddress,
	)
}

func InitConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println(".env not found or not loaded")
	}

	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", "", "Server address (e.g. :8080)")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system base URL")
	flag.Parse()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	} else if cfg.RunAddress == "" {
		cfg.RunAddress = "localhost:8080"
	}

	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	} else if cfg.DatabaseURI == "" {
		return nil, fmt.Errorf("DATABASE_URI is required (set via -d flag or DATABASE_URI env/.env)")
	}

	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envAccrualSystemAddress
	} else if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = ""
	}

	if envAuthSecret := os.Getenv("AUTH_SECRET"); envAuthSecret != "" {
		cfg.AuthSecret = envAuthSecret
	} else if cfg.AuthSecret == "" {
		cfg.AuthSecret = "default-secret-do-not-use-in-production"
	}

	return &cfg, nil
}

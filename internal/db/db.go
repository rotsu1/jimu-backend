package db

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rotsu1/jimu-backend"
)

func InitDB() (*pgxpool.Pool, error) {
	// 1. Check for a single DATABASE_URL first (Standard for Cloud/Production)
	_ = godotenv.Load()
	connString := os.Getenv("DATABASE_URL")

	// 2. Fallback to your individual env vars (Great for Local Docker)
	if connString == "" {
		connString = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)
	}

	// Run Migrations First
	if err := runMigrations(connString); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// 3. Configure the Pool
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

func runMigrations(connString string) error {
	d, err := iofs.New(jimu.MigrationFiles, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, connString)
	if err != nil {
		return err
	}

	// This applies all 'up' migrations that haven't been run yet
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

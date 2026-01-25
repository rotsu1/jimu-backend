package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("Failed to open standard sql db: %v", err)
	}
	defer db.Close()

	migrations := &migrate.FileMigrationSource{
		Dir: "../../migrations",
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		t.Fatalf("Could not apply migrations: %v", err)
	}
	t.Logf("Applied %d migrations!", n)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to create pgxpool: %v", err)
	}

	cleanupTables(t, pool)

	return pool
}

func cleanupTables(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	const query = `TRUNCATE TABLE public.profiles, public.user_identities RESTART IDENTITY CASCADE`

	_, err := db.Exec(context.Background(), query)
	if err != nil {
		t.Fatalf("Failed to cleanup tables: %v", err)
	}
}

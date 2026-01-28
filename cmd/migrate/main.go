package main

import (
	"log"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rotsu1/jimu-backend/internal/db"

	migrate "github.com/rubenv/sql-migrate"
)

func main() {
	pool, err := db.InitDB()
	if err != nil {
		log.Fatalf("failed to initialize DB: %v", err)
	}
	defer pool.Close()

	dbSql := stdlib.OpenDBFromPool(pool)

	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(dbSql, "postgres", migrations, migrate.Up)
	if err != nil {
		panic(err)
	}

	println("Applied", n, "migrations successfully! 🚀")
}

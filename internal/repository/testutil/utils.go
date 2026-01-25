package testutil

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertProfile(ctx context.Context, db *pgxpool.Pool, username string) (uuid.UUID, time.Time, error) {
	var id uuid.UUID
	var updatedAt time.Time
	err := db.QueryRow(
		ctx,
		"INSERT INTO profiles (username) VALUES ($1) RETURNING id, updated_at",
		username,
	).Scan(&id, &updatedAt)

	return id, updatedAt, err
}

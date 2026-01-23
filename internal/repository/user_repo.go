package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) UpsertGoogleUser(ctx context.Context, googleID string, email string) (*models.Profile, error) {
	// Start a Transaction (Tx) to ensure atomicity
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Check if the user already exists
	var existingProfileID uuid.UUID
	err = tx.QueryRow(ctx, "SELECT id FROM profiles WHERE primary_email = $1", email).Scan(&existingProfileID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	var userID uuid.UUID

	query := `
    INSERT INTO user_identities (provider_name, provider_user_id, provider_email)
    VALUES ('google', $1, $2)
    ON CONFLICT (provider_name, provider_user_id) 
    DO UPDATE SET last_sign_in_at = now()
    RETURNING user_id;
	`
	// We scan the resulting user_id into our userID variable
	err = tx.QueryRow(ctx, query, googleID, email).Scan(&userID)
	if err != nil {
		// If the identity already existed, the INSERT above returns nothing.
		// We handle that by fetching the existing ID.
		err = r.DB.QueryRow(ctx, "SELECT user_id FROM user_identities WHERE provider_user_id = $1", googleID).Scan(&userID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.Profile{ID: userID}, nil
}

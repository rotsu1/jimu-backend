package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type BlockedUserRepository struct {
	DB *pgxpool.Pool
}

func NewBlockedUserRepository(db *pgxpool.Pool) *BlockedUserRepository {
	return &BlockedUserRepository{
		DB: db,
	}
}

// Block adds a user to the blocked list. Idempotent operation.
func (r *BlockedUserRepository) Block(
	ctx context.Context,
	blockerID uuid.UUID,
	blockedID uuid.UUID,
) (*models.BlockedUser, error) {
	var blocked models.BlockedUser

	err := r.DB.QueryRow(ctx, insertBlockedUserQuery, blockerID, blockedID).Scan(
		&blocked.BlockerID,
		&blocked.BlockedID,
		&blocked.CreatedAt,
	)
	if err != nil {
		// ON CONFLICT DO NOTHING returns no rows, so we need to handle this
		if errors.Is(err, pgx.ErrNoRows) {
			// Already blocked, fetch existing record
			return r.GetBlockedUser(ctx, blockerID, blockedID)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // foreign key violation
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to block user: %w", err)
	}

	return &blocked, nil
}

// Unblock removes a user from the blocked list.
func (r *BlockedUserRepository) Unblock(
	ctx context.Context,
	blockerID uuid.UUID,
	blockedID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteBlockedUserQuery, blockerID, blockedID)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrBlockedUserNotFound
	}

	return nil
}

// GetBlockedUser gets a specific block relationship.
func (r *BlockedUserRepository) GetBlockedUser(
	ctx context.Context,
	blockerID uuid.UUID,
	blockedID uuid.UUID,
) (*models.BlockedUser, error) {
	var blocked models.BlockedUser

	err := r.DB.QueryRow(ctx, getBlockedUserQuery, blockerID, blockedID).Scan(
		&blocked.BlockerID,
		&blocked.BlockedID,
		&blocked.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBlockedUserNotFound
		}
		return nil, fmt.Errorf("failed to get blocked user: %w", err)
	}

	return &blocked, nil
}

// IsBlocked checks if there's any block relationship between two users (either direction).
func (r *BlockedUserRepository) IsBlocked(
	ctx context.Context,
	userID1 uuid.UUID,
	userID2 uuid.UUID,
) (bool, error) {
	var isBlocked bool
	err := r.DB.QueryRow(ctx, isBlockedQuery, userID1, userID2).Scan(&isBlocked)
	if err != nil {
		return false, fmt.Errorf("failed to check if blocked: %w", err)
	}
	return isBlocked, nil
}

// GetBlockedUsers gets all users blocked by a specific user.
func (r *BlockedUserRepository) GetBlockedUsers(
	ctx context.Context,
	blockerID uuid.UUID,
) ([]*models.BlockedUser, error) {
	rows, err := r.DB.Query(ctx, getBlockedUsersByBlockerIDQuery, blockerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocked users: %w", err)
	}
	defer rows.Close()

	var blockedUsers []*models.BlockedUser
	for rows.Next() {
		var b models.BlockedUser
		err := rows.Scan(
			&b.BlockerID,
			&b.BlockedID,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blocked user: %w", err)
		}
		blockedUsers = append(blockedUsers, &b)
	}

	return blockedUsers, nil
}

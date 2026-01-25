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

type FollowRepository struct {
	DB *pgxpool.Pool
}

func NewFollowRepository(db *pgxpool.Pool) *FollowRepository {
	return &FollowRepository{
		DB: db,
	}
}

// Follow creates a follow request. Status should be 'pending' for private accounts or 'accepted' for public.
func (r *FollowRepository) Follow(
	ctx context.Context,
	followerID uuid.UUID,
	followingID uuid.UUID,
) (*models.Follow, error) {
	var follow models.Follow

	err := r.DB.QueryRow(ctx, upsertFollowQuery, followerID, followingID).Scan(
		&follow.FollowerID,
		&follow.FollowingID,
		&follow.Status,
		&follow.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			case "P0001":
				return nil, ErrBlocked
			}
		}
		return nil, fmt.Errorf("failed to follow: %w", err)
	}

	return &follow, nil
}

// Unfollow removes a follow relationship.
func (r *FollowRepository) Unfollow(
	ctx context.Context,
	followerID uuid.UUID,
	followingID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteFollowQuery, followerID, followingID)
	if err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrFollowNotFound
	}

	return nil
}

// GetFollowStatus gets the follow relationship between two users.
func (r *FollowRepository) GetFollowStatus(
	ctx context.Context,
	followerID uuid.UUID,
	followingID uuid.UUID,
) (*models.Follow, error) {
	var follow models.Follow

	err := r.DB.QueryRow(ctx, getFollowQuery, followerID, followingID).Scan(
		&follow.FollowerID,
		&follow.FollowingID,
		&follow.Status,
		&follow.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFollowNotFound
		}
		return nil, fmt.Errorf("failed to get follow status: %w", err)
	}

	return &follow, nil
}

func (r *FollowRepository) AcceptFollow(
	ctx context.Context,
	followerID uuid.UUID,
	followingID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(
		ctx,
		updateFollowStatusQuery,
		followerID,
		followingID,
	)

	if err != nil {
		return fmt.Errorf("failed to accept follow execution: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrFollowNotFound
	}

	return nil
}

// GetFollowers gets users who follow the specified user.
func (r *FollowRepository) GetFollowers(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]*models.Follow, error) {
	rows, err := r.DB.Query(ctx, getFollowersByUserIDQuery, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()

	var follows []*models.Follow
	for rows.Next() {
		var f models.Follow
		err := rows.Scan(
			&f.FollowerID,
			&f.FollowingID,
			&f.Status,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follow: %w", err)
		}
		follows = append(follows, &f)
	}

	return follows, nil
}

// GetFollowing gets users that the specified user follows.
func (r *FollowRepository) GetFollowing(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]*models.Follow, error) {
	rows, err := r.DB.Query(ctx, getFollowingByUserIDQuery, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}
	defer rows.Close()

	var follows []*models.Follow
	for rows.Next() {
		var f models.Follow
		err := rows.Scan(
			&f.FollowerID,
			&f.FollowingID,
			&f.Status,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follow: %w", err)
		}
		follows = append(follows, &f)
	}

	return follows, nil
}

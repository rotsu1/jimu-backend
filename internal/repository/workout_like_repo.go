package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type WorkoutLikeRepository struct {
	DB *pgxpool.Pool
}

func NewWorkoutLikeRepository(db *pgxpool.Pool) *WorkoutLikeRepository {
	return &WorkoutLikeRepository{
		DB: db,
	}
}

// Like adds a like to a workout. Idempotent operation.
func (r *WorkoutLikeRepository) LikeWorkout(
	ctx context.Context,
	userID uuid.UUID,
	workoutID uuid.UUID,
) (*models.WorkoutLike, error) {
	var like models.WorkoutLike

	err := r.DB.QueryRow(ctx, insertWorkoutLikeQuery, userID, workoutID).Scan(
		&like.UserID,
		&like.WorkoutID,
		&like.CreatedAt,
	)
	if err != nil {
		// ON CONFLICT DO NOTHING returns no rows, so we need to handle this
		if errors.Is(err, pgx.ErrNoRows) {
			// 1. Try to fetch the existing like
			existingLike, getErr := r.GetWorkoutLikeByID(ctx, userID, workoutID)
			if getErr != nil {
				if errors.Is(getErr, ErrWorkoutLikeNotFound) {
					// 2. If GetLike also returns nothing, it means either:
					//    - They aren't allowed to see it (Block/Private)
					//    - The workout doesn't exist
					return nil, ErrWorkoutInteractionNotAllowed
				}
				return nil, getErr
			}
			return existingLike, nil
		}
		return nil, fmt.Errorf("failed to like workout: %w", err)
	}

	return &like, nil
}

// Unlike removes a like from a workout.
func (r *WorkoutLikeRepository) UnlikeWorkout(
	ctx context.Context,
	userID uuid.UUID,
	workoutID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteWorkoutLikeQuery, userID, workoutID)
	if err != nil {
		return fmt.Errorf("failed to unlike workout: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkoutLikeNotFound
	}

	return nil
}

// GetLike gets a specific like.
func (r *WorkoutLikeRepository) GetWorkoutLikeByID(
	ctx context.Context,
	userID uuid.UUID,
	workoutID uuid.UUID,
) (*models.WorkoutLike, error) {
	var like models.WorkoutLike

	err := r.DB.QueryRow(ctx, getWorkoutLikeQuery, userID, workoutID).Scan(
		&like.UserID,
		&like.WorkoutID,
		&like.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkoutLikeNotFound
		}
		return nil, fmt.Errorf("failed to get like: %w", err)
	}

	return &like, nil
}

// GetLikesByWorkoutID gets all likes for a workout.
func (r *WorkoutLikeRepository) GetWorkoutLikesByWorkoutID(
	ctx context.Context,
	workoutID uuid.UUID,
	viewerID uuid.UUID,
	limit int,
	offset int,
) ([]*models.WorkoutLikeDetail, error) {
	// We use Query because we expect 0 or more rows
	rows, err := r.DB.Query(ctx, getLikesByWorkoutIDQuery, workoutID, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workout likes: %w", err)
	}
	defer rows.Close()

	// Initialize as an empty slice so JSON returns [] instead of null
	likes := []*models.WorkoutLikeDetail{}

	for rows.Next() {
		var l models.WorkoutLikeDetail
		err := rows.Scan(
			&l.UserID,
			&l.WorkoutID,
			&l.CreatedAt,
			&l.Username,  // From the Profiles JOIN
			&l.AvatarURL, // From the Profiles JOIN
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout like detail: %w", err)
		}
		likes = append(likes, &l)
	}

	// Always check rows.Err() after the loop!
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return likes, nil
}

// IsLiked checks if a user has liked a workout.
func (r *WorkoutLikeRepository) IsWorkoutLiked(
	ctx context.Context,
	userID uuid.UUID,
	workoutID uuid.UUID,
) (bool, error) {
	var isLiked bool
	err := r.DB.QueryRow(ctx, isWorkoutLikedQuery, userID, workoutID).Scan(&isLiked)
	if err != nil {
		return false, fmt.Errorf("failed to check if liked: %w", err)
	}
	return isLiked, nil
}

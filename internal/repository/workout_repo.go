package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type WorkoutRepository struct {
	DB *pgxpool.Pool
}

func NewWorkoutRepository(db *pgxpool.Pool) *WorkoutRepository {
	return &WorkoutRepository{
		DB: db,
	}
}

func (r *WorkoutRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	name *string,
	comment *string,
	startedAt time.Time,
	endedAt time.Time,
	durationSeconds int,
) (*models.Workout, error) {
	var workout models.Workout

	err := r.DB.QueryRow(
		ctx,
		insertWorkoutQuery,
		userID,
		name,
		comment,
		startedAt,
		endedAt,
		durationSeconds,
	).Scan(
		&workout.ID,
		&workout.UserID,
		&workout.Name,
		&workout.Comment,
		&workout.StartedAt,
		&workout.EndedAt,
		&workout.DurationSeconds,
		&workout.TotalWeight,
		&workout.LikesCount,
		&workout.CommentsCount,
		&workout.CreatedAt,
		&workout.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to create workout: %w", err)
	}

	return &workout, nil
}

func (r *WorkoutRepository) GetWorkoutByID(
	ctx context.Context,
	workoutID uuid.UUID,
	viewerID uuid.UUID,
) (*models.Workout, error) {
	var workout models.Workout

	err := r.DB.QueryRow(ctx, getWorkoutByIDQuery, workoutID, viewerID).Scan(
		&workout.ID,
		&workout.UserID,
		&workout.Name,
		&workout.Comment,
		&workout.StartedAt,
		&workout.EndedAt,
		&workout.DurationSeconds,
		&workout.TotalWeight,
		&workout.LikesCount,
		&workout.CommentsCount,
		&workout.CreatedAt,
		&workout.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkoutNotFound
		}
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	return &workout, nil
}

func (r *WorkoutRepository) GetWorkoutsByUserID(
	ctx context.Context,
	targetID uuid.UUID,
	viewerID uuid.UUID,
	limit int,
	offset int,
) ([]*models.Workout, error) {
	rows, err := r.DB.Query(ctx, getWorkoutsByUserIDQuery, targetID, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get workouts: %w", err)
	}
	defer rows.Close()

	var workouts []*models.Workout
	for rows.Next() {
		var workout models.Workout
		err := rows.Scan(
			&workout.ID,
			&workout.UserID,
			&workout.Name,
			&workout.Comment,
			&workout.StartedAt,
			&workout.EndedAt,
			&workout.DurationSeconds,
			&workout.TotalWeight,
			&workout.LikesCount,
			&workout.CommentsCount,
			&workout.CreatedAt,
			&workout.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout: %w", err)
		}
		workouts = append(workouts, &workout)
	}

	return workouts, nil
}

func (r *WorkoutRepository) UpdateWorkout(
	ctx context.Context,
	id uuid.UUID,
	updates models.UpdateWorkoutRequest,
) error {
	var sets []string
	var args []interface{}
	i := 1

	if updates.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		if *updates.Name == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Name)
		}
		i++
	}
	if updates.Comment != nil {
		sets = append(sets, fmt.Sprintf("comment = $%d", i))
		if *updates.Comment == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Comment)
		}
		i++
	}
	if updates.EndedAt != nil {
		sets = append(sets, fmt.Sprintf("ended_at = $%d", i))
		args = append(args, *updates.EndedAt)
		i++
	}
	if updates.DurationSeconds != nil {
		sets = append(sets, fmt.Sprintf("duration_seconds = $%d", i))
		args = append(args, *updates.DurationSeconds)
		i++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(
		"UPDATE workouts SET %s WHERE id = $%d",
		strings.Join(sets, ", "),
		i,
	)
	args = append(args, id)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrWorkoutNotFound
	}

	return nil
}

func (r *WorkoutRepository) DeleteWorkout(ctx context.Context, id uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, deleteWorkoutByIDQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkoutNotFound
	}

	return nil
}

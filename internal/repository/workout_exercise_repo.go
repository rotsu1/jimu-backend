package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type WorkoutExerciseRepository struct {
	DB *pgxpool.Pool
}

func NewWorkoutExerciseRepository(db *pgxpool.Pool) *WorkoutExerciseRepository {
	return &WorkoutExerciseRepository{
		DB: db,
	}
}

func (r *WorkoutExerciseRepository) CreateWorkoutExercise(
	ctx context.Context,
	workoutID uuid.UUID,
	exerciseID uuid.UUID,
	orderIndex *int,
	memo *string,
	restTimerSeconds *int,
	userID uuid.UUID,
) (*models.WorkoutExercise, error) {
	var we models.WorkoutExercise

	err := r.DB.QueryRow(
		ctx,
		insertWorkoutExerciseQuery,
		workoutID,
		exerciseID,
		orderIndex,
		memo,
		restTimerSeconds,
		userID,
	).Scan(
		&we.ID,
		&we.WorkoutID,
		&we.ExerciseID,
		&we.OrderIndex,
		&we.Memo,
		&we.RestTimerSeconds,
		&we.CreatedAt,
		&we.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReferenceViolation
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // Unique Violation
				return nil, ErrAlreadyExists
			case "23503": // Foreign Key Violation
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to create workout set: %w", err)
	}

	return &we, nil
}

func (r *WorkoutExerciseRepository) GetWorkoutExerciseByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.WorkoutExercise, error) {
	var we models.WorkoutExercise

	err := r.DB.QueryRow(ctx, getWorkoutExerciseByIDQuery, id).Scan(
		&we.ID,
		&we.WorkoutID,
		&we.ExerciseID,
		&we.OrderIndex,
		&we.Memo,
		&we.RestTimerSeconds,
		&we.CreatedAt,
		&we.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkoutExerciseNotFound
		}
		return nil, fmt.Errorf("failed to get workout exercise: %w", err)
	}

	return &we, nil
}

func (r *WorkoutExerciseRepository) GetWorkoutExercisesByWorkoutID(
	ctx context.Context,
	workoutID uuid.UUID,
) ([]*models.WorkoutExercise, error) {
	rows, err := r.DB.Query(ctx, getWorkoutExercisesByWorkoutIDQuery, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout exercises: %w", err)
	}
	defer rows.Close()

	var exercises []*models.WorkoutExercise
	for rows.Next() {
		var we models.WorkoutExercise
		err := rows.Scan(
			&we.ID,
			&we.WorkoutID,
			&we.ExerciseID,
			&we.OrderIndex,
			&we.Memo,
			&we.RestTimerSeconds,
			&we.CreatedAt,
			&we.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout exercise: %w", err)
		}
		exercises = append(exercises, &we)
	}

	return exercises, nil
}

func (r *WorkoutExerciseRepository) UpdateWorkoutExercise(
	ctx context.Context,
	workoutExerciseID uuid.UUID,
	updates models.UpdateWorkoutExerciseRequest,
	userID uuid.UUID,
) error {
	var sets []string
	var args []interface{}
	i := 1

	if updates.OrderIndex != nil {
		sets = append(sets, fmt.Sprintf("order_index = $%d", i))
		if *updates.OrderIndex == 0 {
			args = append(args, nil)
		} else {
			args = append(args, *updates.OrderIndex)
		}
		i++
	}
	if updates.Memo != nil {
		sets = append(sets, fmt.Sprintf("memo = $%d", i))
		if *updates.Memo == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Memo)
		}
		i++
	}
	if updates.RestTimerSeconds != nil {
		sets = append(sets, fmt.Sprintf("rest_timer_seconds = $%d", i))
		if *updates.RestTimerSeconds == 0 {
			args = append(args, nil)
		} else {
			args = append(args, *updates.RestTimerSeconds)
		}
		i++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
    UPDATE public.workout_exercises 
    SET %s 
    WHERE id = $%d 
    AND workout_id IN (
        SELECT id FROM public.workouts 
        WHERE (user_id = $%d OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $%d))
    )`,
		strings.Join(sets, ", "),
		i,   // The Exercise ID
		i+1, // The User ID (Owner check)
		i+2, // The User ID (Admin check)
	)

	// Append the IDs to the args slice
	args = append(args, workoutExerciseID, userID, userID)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workout exercise: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrWorkoutExerciseNotFound
	}

	return nil
}

func (r *WorkoutExerciseRepository) DeleteWorkoutExercise(
	ctx context.Context,
	workoutExerciseID uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteWorkoutExerciseByIDQuery, workoutExerciseID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete workout exercise: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkoutExerciseNotFound
	}

	return nil
}

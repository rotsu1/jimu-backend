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

type WorkoutSetRepository struct {
	DB *pgxpool.Pool
}

func NewWorkoutSetRepository(db *pgxpool.Pool) *WorkoutSetRepository {
	return &WorkoutSetRepository{
		DB: db,
	}
}

func (r *WorkoutSetRepository) CreateWorkoutSet(
	ctx context.Context,
	workoutExerciseID uuid.UUID,
	weight *float64,
	reps *int,
	isCompleted bool,
	orderIndex int,
	userID uuid.UUID,
) (*models.WorkoutSet, error) {
	var ws models.WorkoutSet

	err := r.DB.QueryRow(
		ctx, insertWorkoutSetQuery,
		workoutExerciseID,
		weight,
		reps,
		isCompleted,
		orderIndex,
		userID,
	).Scan(
		&ws.ID,
		&ws.WorkoutExerciseID,
		&ws.Weight,
		&ws.Reps,
		&ws.IsCompleted,
		&ws.OrderIndex,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if err != nil {
		// 1. Check for NoRows (The Guard blocked the insert)
		if errors.Is(err, pgx.ErrNoRows) {
			// We return ErrReferenceViolation or ErrNotFound here
			// because the parent exercise wasn't found/owned.
			return nil, ErrReferenceViolation
		}

		// 2. Check for other DB errors using your helper
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

	return &ws, nil
}

func (r *WorkoutSetRepository) GetWorkoutSetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.WorkoutSet, error) {
	var ws models.WorkoutSet

	err := r.DB.QueryRow(ctx, getWorkoutSetByIDQuery, id).Scan(
		&ws.ID,
		&ws.WorkoutExerciseID,
		&ws.Weight,
		&ws.Reps,
		&ws.IsCompleted,
		&ws.OrderIndex,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkoutSetNotFound
		}
		return nil, fmt.Errorf("failed to get workout set: %w", err)
	}

	return &ws, nil
}

func (r *WorkoutSetRepository) GetWorkoutSetsByWorkoutExerciseID(
	ctx context.Context,
	workoutExerciseID uuid.UUID,
) ([]*models.WorkoutSet, error) {
	rows, err := r.DB.Query(ctx, getWorkoutSetsByWorkoutExerciseIDQuery, workoutExerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout sets: %w", err)
	}
	defer rows.Close()

	var sets []*models.WorkoutSet
	for rows.Next() {
		var ws models.WorkoutSet
		err := rows.Scan(
			&ws.ID,
			&ws.WorkoutExerciseID,
			&ws.Weight,
			&ws.Reps,
			&ws.IsCompleted,
			&ws.OrderIndex,
			&ws.CreatedAt,
			&ws.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout set: %w", err)
		}
		sets = append(sets, &ws)
	}

	return sets, nil
}

func (r *WorkoutSetRepository) UpdateWorkoutSet(
	ctx context.Context,
	workoutSetID uuid.UUID,
	userID uuid.UUID,
	updates models.UpdateWorkoutSetRequest,
) error {
	var sets []string
	var args []interface{}
	i := 1

	if updates.Weight != nil {
		sets = append(sets, fmt.Sprintf("weight = $%d", i))
		if *updates.Weight == 0 {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Weight)
		}
		i++
	}
	if updates.Reps != nil {
		sets = append(sets, fmt.Sprintf("reps = $%d", i))
		if *updates.Reps == 0 {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Reps)
		}
		i++
	}
	if updates.OrderIndex != nil {
		sets = append(sets, fmt.Sprintf("order_index = $%d", i))
		if *updates.OrderIndex == 0 {
			args = append(args, nil)
		} else {
			args = append(args, *updates.OrderIndex)
		}
		i++
	}

	if len(sets) == 0 {
		return nil // Guard: return immediately if nothing to update
	}

	// Use a JOIN or subquery in the UPDATE to verify ownership
	query := fmt.Sprintf(`
    UPDATE public.workout_sets 
    SET %s, updated_at = NOW() 
    WHERE id = $%d 
    AND workout_exercise_id IN (
        SELECT we.id 
        FROM public.workout_exercises we
        JOIN public.workouts w ON we.workout_id = w.id
        WHERE (w.user_id = $%d OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $%d))
    )`,
		strings.Join(sets, ", "),
		i,   // workoutSetID
		i+1, // userID
		i+2, // userID again for Admin bypass
	)

	args = append(args, workoutSetID, userID, userID)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // Unique Violation
				return ErrAlreadyExists
			case "23503": // Foreign Key Violation
				return ErrReferenceViolation
			}
		}

		return fmt.Errorf("failed to update workout set: %w", err)
	}

	if res.RowsAffected() == 0 {
		// If 0 rows are affected, it either doesn't exist OR the user isn't the owner
		return ErrWorkoutSetNotFound
	}

	return nil
}

func (r *WorkoutSetRepository) DeleteWorkoutSet(ctx context.Context, workoutSetID uuid.UUID, userID uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, deleteWorkoutSetByIDQuery, workoutSetID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete workout set: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkoutSetNotFound
	}

	return nil
}

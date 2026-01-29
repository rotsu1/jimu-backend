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

type RoutineExerciseRepository struct {
	DB *pgxpool.Pool
}

func NewRoutineExerciseRepository(db *pgxpool.Pool) *RoutineExerciseRepository {
	return &RoutineExerciseRepository{
		DB: db,
	}
}

func (r *RoutineExerciseRepository) CreateRoutineExercise(
	ctx context.Context,
	routineID uuid.UUID,
	exerciseID uuid.UUID,
	orderIndex int,
	restTimerSeconds *int,
	memo *string,
	userID uuid.UUID,
) (*models.RoutineExercise, error) {
	var re models.RoutineExercise

	// userID passed as $6 for ownership check
	err := r.DB.QueryRow(ctx, insertRoutineExerciseQuery, routineID, exerciseID, orderIndex, restTimerSeconds, memo, userID).Scan(
		&re.ID,
		&re.RoutineID,
		&re.ExerciseID,
		&re.OrderIndex,
		&re.RestTimerSeconds,
		&re.Memo,
		&re.CreatedAt,
		&re.UpdatedAt,
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

	return &re, nil
}

func (r *RoutineExerciseRepository) GetRoutineExerciseByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.RoutineExercise, error) {
	var re models.RoutineExercise

	err := r.DB.QueryRow(ctx, getRoutineExerciseByIDQuery, id).Scan(
		&re.ID,
		&re.RoutineID,
		&re.ExerciseID,
		&re.OrderIndex,
		&re.RestTimerSeconds,
		&re.Memo,
		&re.CreatedAt,
		&re.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoutineExerciseNotFound
		}
		return nil, fmt.Errorf("failed to get routine exercise: %w", err)
	}

	return &re, nil
}

func (r *RoutineExerciseRepository) GetRoutineExercisesByRoutineID(
	ctx context.Context,
	routineID uuid.UUID,
) ([]*models.RoutineExercise, error) {
	rows, err := r.DB.Query(ctx, getRoutineExercisesByRoutineIDQuery, routineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routine exercises: %w", err)
	}
	defer rows.Close()

	var exercises []*models.RoutineExercise
	for rows.Next() {
		var re models.RoutineExercise
		err := rows.Scan(
			&re.ID,
			&re.RoutineID,
			&re.ExerciseID,
			&re.OrderIndex,
			&re.RestTimerSeconds,
			&re.Memo,
			&re.CreatedAt,
			&re.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan routine exercise: %w", err)
		}
		exercises = append(exercises, &re)
	}

	return exercises, nil
}

func (r *RoutineExerciseRepository) UpdateRoutineExercise(
	ctx context.Context,
	id uuid.UUID,
	updates models.UpdateRoutineExerciseRequest,
	userID uuid.UUID,
) error {
	var sets []string
	var args []interface{}
	i := 1

	if updates.OrderIndex != nil {
		sets = append(sets, fmt.Sprintf("order_index = $%d", i))
		args = append(args, *updates.OrderIndex)
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
	if updates.Memo != nil {
		sets = append(sets, fmt.Sprintf("memo = $%d", i))
		if *updates.Memo == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Memo)
		}
		i++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
    UPDATE public.routine_exercises 
    SET %s 
    WHERE id = $%d 
    AND routine_id IN (
        SELECT id FROM public.routines 
        WHERE (user_id = $%d OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $%d))
    )`,
		strings.Join(sets, ", "),
		i,   // The Routine Exercise ID
		i+1, // UserID for Ownership check
		i+2, // UserID for Admin check
	)

	args = append(args, id, userID, userID)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update routine exercise: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrRoutineExerciseNotFound
	}

	return nil
}

func (r *RoutineExerciseRepository) DeleteRoutineExercise(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteRoutineExerciseByIDQuery, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete routine exercise: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrRoutineExerciseNotFound
	}

	return nil
}

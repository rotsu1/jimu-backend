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

type RoutineSetRepository struct {
	DB *pgxpool.Pool
}

func NewRoutineSetRepository(db *pgxpool.Pool) *RoutineSetRepository {
	return &RoutineSetRepository{
		DB: db,
	}
}

func (r *RoutineSetRepository) CreateRoutineSet(
	ctx context.Context,
	routineExerciseID uuid.UUID,
	weight *float64,
	reps *int,
	orderIndex *int,
	userID uuid.UUID,
) (*models.RoutineSet, error) {
	var rs models.RoutineSet

	err := r.DB.QueryRow(ctx, insertRoutineSetQuery, routineExerciseID, weight, reps, orderIndex, userID).Scan(
		&rs.ID,
		&rs.RoutineExerciseID,
		&rs.Weight,
		&rs.Reps,
		&rs.OrderIndex,
		&rs.CreatedAt,
		&rs.UpdatedAt,
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

	return &rs, nil
}

func (r *RoutineSetRepository) GetRoutineSetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.RoutineSet, error) {
	var rs models.RoutineSet

	err := r.DB.QueryRow(ctx, getRoutineSetByIDQuery, id).Scan(
		&rs.ID,
		&rs.RoutineExerciseID,
		&rs.Weight,
		&rs.Reps,
		&rs.OrderIndex,
		&rs.CreatedAt,
		&rs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoutineSetNotFound
		}
		return nil, fmt.Errorf("failed to get routine set: %w", err)
	}

	return &rs, nil
}

func (r *RoutineSetRepository) GetRoutineSetsByRoutineExerciseID(
	ctx context.Context,
	routineExerciseID uuid.UUID,
) ([]*models.RoutineSet, error) {
	rows, err := r.DB.Query(ctx, getRoutineSetsByRoutineExerciseIDQuery, routineExerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routine sets: %w", err)
	}
	defer rows.Close()

	var sets []*models.RoutineSet
	for rows.Next() {
		var rs models.RoutineSet
		err := rows.Scan(
			&rs.ID,
			&rs.RoutineExerciseID,
			&rs.Weight,
			&rs.Reps,
			&rs.OrderIndex,
			&rs.CreatedAt,
			&rs.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan routine set: %w", err)
		}
		sets = append(sets, &rs)
	}

	return sets, nil
}

func (r *RoutineSetRepository) UpdateRoutineSet(
	ctx context.Context,
	id uuid.UUID,
	updates models.UpdateRoutineSetRequest,
	userID uuid.UUID,
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
		return nil
	}

	query := fmt.Sprintf(`
    UPDATE public.routine_sets 
    SET %s 
    WHERE id = $%d 
      AND routine_exercise_id IN (
        SELECT re.id 
        FROM public.routine_exercises re
        JOIN public.routines r ON re.routine_id = r.id
        WHERE (r.user_id = $%d OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $%d))
      )`,
		strings.Join(sets, ", "),
		i,   // The Set ID
		i+1, // The User ID for Ownership
		i+2, // The User ID for Admin Bypass
	)

	// Append args in the correct order to match the placeholders
	args = append(args, id, userID, userID)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update routine set: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrRoutineSetNotFound
	}

	return nil
}

func (r *RoutineSetRepository) DeleteRoutineSet(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, deleteRoutineSetByIDQuery, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete routine set: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrRoutineSetNotFound
	}

	return nil
}

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

type ExerciseRepository struct {
	DB *pgxpool.Pool
}

func NewExerciseRepository(db *pgxpool.Pool) *ExerciseRepository {
	return &ExerciseRepository{
		DB: db,
	}
}

func (r *ExerciseRepository) CreateExercise(
	ctx context.Context,
	userID *uuid.UUID,
	name string,
	suggestedRestSeconds *int,
	icon *string,
	requesterID uuid.UUID,
) (*models.Exercise, error) {
	var exercise models.Exercise

	err := r.DB.QueryRow(ctx, insertExerciseQuery, userID, name, suggestedRestSeconds, icon, requesterID).Scan(
		&exercise.ID,
		&exercise.UserID,
		&exercise.Name,
		&exercise.SuggestedRestSeconds,
		&exercise.Icon,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
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

	return &exercise, nil
}

func (r *ExerciseRepository) GetExerciseByID(
	ctx context.Context,
	exerciseID uuid.UUID,
	userID uuid.UUID,
) (*models.Exercise, error) {
	var exercise models.Exercise

	err := r.DB.QueryRow(ctx, getExerciseByIDQuery, exerciseID, userID).Scan(
		&exercise.ID,
		&exercise.UserID,
		&exercise.Name,
		&exercise.SuggestedRestSeconds,
		&exercise.Icon,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrExerciseNotFound
		}
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	return &exercise, nil
}

func (r *ExerciseRepository) GetExercisesByUserID(
	ctx context.Context,
	viewerID uuid.UUID,
	targetID uuid.UUID,
) ([]*models.Exercise, error) {
	rows, err := r.DB.Query(ctx, getExercisesByUserIDQuery, viewerID, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercises: %w", err)
	}
	defer rows.Close()

	var exercises []*models.Exercise
	for rows.Next() {
		var exercise models.Exercise
		err := rows.Scan(
			&exercise.ID,
			&exercise.UserID,
			&exercise.Name,
			&exercise.SuggestedRestSeconds,
			&exercise.Icon,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exercise: %w", err)
		}
		exercises = append(exercises, &exercise)
	}

	return exercises, nil
}

func (r *ExerciseRepository) UpdateExercise(
	ctx context.Context,
	id uuid.UUID,
	updates models.UpdateExerciseRequest,
	userID uuid.UUID,
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
	if updates.SuggestedRestSeconds != nil {
		sets = append(sets, fmt.Sprintf("suggested_rest_seconds = $%d", i))
		if *updates.SuggestedRestSeconds == 0 {
			args = append(args, nil)
		} else {
			args = append(args, *updates.SuggestedRestSeconds)
		}
		i++
	}
	if updates.Icon != nil {
		sets = append(sets, fmt.Sprintf("icon = $%d", i))
		if *updates.Icon == "" {
			args = append(args, nil)
		} else {
			args = append(args, *updates.Icon)
		}
		i++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
    UPDATE public.exercises 
    SET %s 
    WHERE id = $%d 
    AND (
        user_id = $%d 
        OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $%d)
    )`,
		strings.Join(sets, ", "),
		i,   // The Exercise ID ($%d)
		i+1, // The User ID for Ownership ($%d)
		i+2, // The User ID for Admin Bypass ($%d)
	)

	args = append(args, id, userID, userID)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update exercise: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrExerciseNotFound
	}

	return nil
}

func (r *ExerciseRepository) DeleteExercise(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteExerciseByIDQuery, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete exercise: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrExerciseNotFound
	}

	return nil
}

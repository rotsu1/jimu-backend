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

type ExerciseTargetMuscleRepository struct {
	DB *pgxpool.Pool
}

func NewExerciseTargetMuscleRepository(db *pgxpool.Pool) *ExerciseTargetMuscleRepository {
	return &ExerciseTargetMuscleRepository{
		DB: db,
	}
}

// GetByExerciseID gets all target muscles for an exercise.
func (r *ExerciseTargetMuscleRepository) GetByExerciseID(
	ctx context.Context,
	exerciseID uuid.UUID,
) ([]*models.ExerciseTargetMuscle, error) {
	rows, err := r.DB.Query(ctx, getExerciseTargetMusclesByExerciseIDQuery, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise target muscles: %w", err)
	}
	defer rows.Close()

	var muscles []*models.ExerciseTargetMuscle
	for rows.Next() {
		var m models.ExerciseTargetMuscle
		err := rows.Scan(
			&m.ExerciseID,
			&m.MuscleID,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exercise target muscle: %w", err)
		}
		muscles = append(muscles, &m)
	}

	return muscles, nil
}

// Add adds a target muscle to an exercise. Idempotent operation.
func (r *ExerciseTargetMuscleRepository) AddTargetMuscle(
	ctx context.Context,
	exerciseID uuid.UUID,
	muscleID uuid.UUID,
	userID uuid.UUID,
) (*models.ExerciseTargetMuscle, error) {
	var etm models.ExerciseTargetMuscle

	err := r.DB.QueryRow(ctx, insertExerciseTargetMuscleQuery, exerciseID, muscleID, userID).Scan(
		&etm.ExerciseID,
		&etm.MuscleID,
		&etm.CreatedAt,
		&etm.UpdatedAt,
	)
	if err != nil {
		// ON CONFLICT DO NOTHING returns no rows
		if errors.Is(err, pgx.ErrNoRows) {
			// Already exists, fetch existing record
			return r.getTargetMuscle(ctx, exerciseID, muscleID)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to add exercise target muscle: %w", err)
	}

	return &etm, nil
}

// get fetches a specific exercise-muscle relationship.
func (r *ExerciseTargetMuscleRepository) getTargetMuscle(
	ctx context.Context,
	exerciseID uuid.UUID,
	muscleID uuid.UUID,
) (*models.ExerciseTargetMuscle, error) {
	var etm models.ExerciseTargetMuscle

	query := `
		SELECT exercise_id, muscle_id, created_at, updated_at
		FROM public.exercise_target_muscles
		WHERE exercise_id = $1 AND muscle_id = $2
	`
	err := r.DB.QueryRow(ctx, query, exerciseID, muscleID).Scan(
		&etm.ExerciseID,
		&etm.MuscleID,
		&etm.CreatedAt,
		&etm.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get exercise target muscle: %w", err)
	}

	return &etm, nil
}

// Set replaces all target muscles for an exercise with the given list.
func (r *ExerciseTargetMuscleRepository) SetTargetMuscles(
	ctx context.Context,
	exerciseID uuid.UUID,
	muscleIDs []uuid.UUID,
	userID uuid.UUID,
) error {
	// Start a transaction
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing
	_, err = tx.Exec(ctx, deleteExerciseTargetMusclesByExerciseIDQuery, exerciseID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete existing target muscles: %w", err)
	}

	// Insert new
	for _, muscleID := range muscleIDs {
		_, err = tx.Exec(ctx, insertExerciseTargetMuscleQuery, exerciseID, muscleID, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23503":
					return ErrReferenceViolation
				}
			}
			return fmt.Errorf("failed to insert target muscle: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Remove removes a target muscle from an exercise.
func (r *ExerciseTargetMuscleRepository) RemoveTargetMuscle(
	ctx context.Context,
	exerciseID uuid.UUID,
	muscleID uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteExerciseTargetMuscleQuery, exerciseID, muscleID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove exercise target muscle: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

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

type WorkoutImageRepository struct {
	DB *pgxpool.Pool
}

func NewWorkoutImageRepository(db *pgxpool.Pool) *WorkoutImageRepository {
	return &WorkoutImageRepository{
		DB: db,
	}
}

func (r *WorkoutImageRepository) CreateWorkoutImage(
	ctx context.Context,
	workoutID uuid.UUID,
	storagePath string,
	displayOrder int,
	userID uuid.UUID,
) (*models.WorkoutImage, error) {
	var wi models.WorkoutImage

	err := r.DB.QueryRow(ctx, insertWorkoutImageQuery, workoutID, storagePath, displayOrder, userID).Scan(
		&wi.ID,
		&wi.WorkoutID,
		&wi.StoragePath,
		&wi.DisplayOrder,
		&wi.CreatedAt,
		&wi.UpdatedAt,
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

		return nil, fmt.Errorf("failed to create workout image: %w", err)
	}

	return &wi, nil
}

func (r *WorkoutImageRepository) GetWorkoutImageByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.WorkoutImage, error) {
	var wi models.WorkoutImage

	err := r.DB.QueryRow(ctx, getWorkoutImageByIDQuery, id).Scan(
		&wi.ID,
		&wi.WorkoutID,
		&wi.StoragePath,
		&wi.DisplayOrder,
		&wi.CreatedAt,
		&wi.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkoutImageNotFound
		}
		return nil, fmt.Errorf("failed to get workout image: %w", err)
	}

	return &wi, nil
}

func (r *WorkoutImageRepository) GetWorkoutImagesByWorkoutID(
	ctx context.Context,
	workoutID uuid.UUID,
) ([]*models.WorkoutImage, error) {
	rows, err := r.DB.Query(ctx, getWorkoutImagesByWorkoutIDQuery, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout images: %w", err)
	}
	defer rows.Close()

	var images []*models.WorkoutImage
	for rows.Next() {
		var wi models.WorkoutImage
		err := rows.Scan(
			&wi.ID,
			&wi.WorkoutID,
			&wi.StoragePath,
			&wi.DisplayOrder,
			&wi.CreatedAt,
			&wi.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workout image: %w", err)
		}
		images = append(images, &wi)
	}

	return images, nil
}

func (r *WorkoutImageRepository) DeleteWorkoutImage(
	ctx context.Context,
	workoutImageID uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteWorkoutImageByIDQuery, workoutImageID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete workout image: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkoutImageNotFound
	}

	return nil
}

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

type MuscleRepository struct {
	DB *pgxpool.Pool
}

func NewMuscleRepository(db *pgxpool.Pool) *MuscleRepository {
	return &MuscleRepository{
		DB: db,
	}
}

// GetAll returns all muscles ordered by name.
func (r *MuscleRepository) GetAllMuscles(ctx context.Context) ([]*models.Muscle, error) {
	rows, err := r.DB.Query(ctx, getAllMusclesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get muscles: %w", err)
	}
	defer rows.Close()

	var muscles []*models.Muscle
	for rows.Next() {
		var m models.Muscle
		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan muscle: %w", err)
		}
		muscles = append(muscles, &m)
	}

	return muscles, nil
}

// GetByID returns a muscle by ID.
func (r *MuscleRepository) GetMuscleByID(ctx context.Context, id uuid.UUID) (*models.Muscle, error) {
	var muscle models.Muscle

	err := r.DB.QueryRow(ctx, getMuscleByIDQuery, id).Scan(
		&muscle.ID,
		&muscle.Name,
		&muscle.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMuscleNotFound
		}
		return nil, fmt.Errorf("failed to get muscle: %w", err)
	}

	return &muscle, nil
}

// GetByName returns a muscle by name.
func (r *MuscleRepository) GetMuscleByName(ctx context.Context, name string) (*models.Muscle, error) {
	var muscle models.Muscle

	err := r.DB.QueryRow(ctx, getMuscleByNameQuery, name).Scan(
		&muscle.ID,
		&muscle.Name,
		&muscle.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMuscleNotFound
		}
		return nil, fmt.Errorf("failed to get muscle: %w", err)
	}

	return &muscle, nil
}

func (r *MuscleRepository) CreateMuscle(
	ctx context.Context,
	name string,
	userID uuid.UUID,
) (*models.Muscle, error) {
	var muscle models.Muscle

	err := r.DB.QueryRow(ctx, createMuscleQuery, name, userID).Scan(
		&muscle.ID,
		&muscle.Name,
		&muscle.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUnauthorizedAction
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
		return nil, fmt.Errorf("failed to create muscle: %w", err)
	}

	return &muscle, nil
}

func (r *MuscleRepository) DeleteMuscle(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID, // The ID of the person trying to delete
) error {
	// Execute the query passing the Muscle ID and the Requester's ID
	res, err := r.DB.Exec(ctx, deleteMuscleQuery, id, userID)
	if err != nil {
		// Check for Foreign Key violations (e.g., this muscle is still linked to an exercise)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // Unique Violation
				return ErrAlreadyExists
			case "23503": // Foreign Key Violation
				return ErrReferenceViolation
			}
			return fmt.Errorf("failed to create muscle: %w", err)
		}

	}

	// If no rows were affected, the ID didn't exist OR the user isn't an admin
	if res.RowsAffected() == 0 {
		return ErrUnauthorizedAction
	}

	return nil
}

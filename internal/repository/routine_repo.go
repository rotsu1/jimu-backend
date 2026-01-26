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

type RoutineRepository struct {
	DB *pgxpool.Pool
}

func NewRoutineRepository(db *pgxpool.Pool) *RoutineRepository {
	return &RoutineRepository{
		DB: db,
	}
}

func (r *RoutineRepository) CreateRoutine(
	ctx context.Context,
	userID uuid.UUID,
	name string,
) (*models.Routine, error) {
	var routine models.Routine

	err := r.DB.QueryRow(ctx, insertRoutineQuery, userID, name).Scan(
		&routine.ID,
		&routine.UserID,
		&routine.Name,
		&routine.CreatedAt,
		&routine.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to create routine: %w", err)
	}

	return &routine, nil
}

func (r *RoutineRepository) GetRoutineByID(
	ctx context.Context,
	id uuid.UUID,
	viewerID uuid.UUID,
) (*models.Routine, error) {
	var routine models.Routine

	err := r.DB.QueryRow(ctx, getRoutineByIDQuery, id, viewerID).Scan(
		&routine.ID,
		&routine.UserID,
		&routine.Name,
		&routine.CreatedAt,
		&routine.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoutineNotFound
		}
		return nil, fmt.Errorf("failed to get routine: %w", err)
	}

	return &routine, nil
}

func (r *RoutineRepository) GetRoutinesByUserID(
	ctx context.Context,
	userID uuid.UUID,
	viewerID uuid.UUID,
) ([]*models.Routine, error) {
	rows, err := r.DB.Query(ctx, getRoutinesByUserIDQuery, userID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routines: %w", err)
	}
	defer rows.Close()

	var routines []*models.Routine
	for rows.Next() {
		var routine models.Routine
		err := rows.Scan(
			&routine.ID,
			&routine.UserID,
			&routine.Name,
			&routine.CreatedAt,
			&routine.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan routine: %w", err)
		}
		routines = append(routines, &routine)
	}

	return routines, nil
}

func (r *RoutineRepository) UpdateRoutine(
	ctx context.Context,
	id uuid.UUID,
	updates models.UpdateRoutineRequest,
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

	if len(sets) == 0 {
		return nil
	}

	// Add owner check to the WHERE clause
	query := fmt.Sprintf(
		"UPDATE routines SET %s WHERE id = $%d AND user_id = $%d",
		strings.Join(sets, ", "),
		i,
		i+1,
	)
	args = append(args, id, userID)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update routine: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrRoutineNotFound
	}

	return nil
}

func (r *RoutineRepository) DeleteRoutine(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	commandTag, err := r.DB.Exec(ctx, deleteRoutineByIDQuery, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrRoutineNotFound
	}

	return nil
}

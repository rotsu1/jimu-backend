package repository

import (
	"context"
	"encoding/json"
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
		"UPDATE workouts SET %s WHERE id = $%d AND (user_id = $%d OR EXISTS (SELECT 1 FROM public.sys_admins WHERE user_id = $%d))",
		strings.Join(sets, ", "),
		i,
		i+1,
		i+2,
	)
	args = append(args, id, userID, userID)

	res, err := r.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrWorkoutNotFound
	}

	return nil
}

func (r *WorkoutRepository) DeleteWorkout(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteWorkoutByIDQuery, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWorkoutNotFound
	}

	return nil
}

func (r *WorkoutRepository) GetTimelineWorkouts(
	ctx context.Context,
	viewerID uuid.UUID, // $2: The person viewing the feed
	targetID uuid.UUID, // $1: The person whose profile we are looking at
	limit int, // $3
	offset int, // $4
) ([]*models.TimelineWorkout, error) {

	// 1. Query with correct Argument Order
	rows, err := r.DB.Query(ctx, getTimelineWorkoutsQuery, targetID, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline workouts: %w", err)
	}
	defer rows.Close()
	return r.scanTimelineWorkoutRows(rows)
}

// GetFollowingTimelineWorkouts returns timeline workouts from the viewer and users they follow (accepted).
func (r *WorkoutRepository) GetFollowingTimelineWorkouts(
	ctx context.Context,
	viewerID uuid.UUID,
	limit int,
	offset int,
) ([]*models.TimelineWorkout, error) {
	rows, err := r.DB.Query(ctx, getFollowingTimelineWorkoutsQuery, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get following timeline workouts: %w", err)
	}
	defer rows.Close()
	return r.scanTimelineWorkoutRows(rows)
}

// GetForYouTimelineWorkouts returns timeline workouts from any visible user, ordered by engagement (likes + comments) then recency.
func (r *WorkoutRepository) GetForYouTimelineWorkouts(
	ctx context.Context,
	viewerID uuid.UUID,
	limit int,
	offset int,
) ([]*models.TimelineWorkout, error) {
	rows, err := r.DB.Query(ctx, getForYouTimelineWorkoutsQuery, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get for-you timeline workouts: %w", err)
	}
	defer rows.Close()
	return r.scanTimelineWorkoutRows(rows)
}

// scanTimelineWorkoutRows scans pgx rows from a timeline workout query into []*models.TimelineWorkout.
func (r *WorkoutRepository) scanTimelineWorkoutRows(rows pgx.Rows) ([]*models.TimelineWorkout, error) {
	workouts := []*models.TimelineWorkout{}
	for rows.Next() {
		var workout models.TimelineWorkout
		var exercisesJSON []byte
		var commentsJSON []byte
		var imagesJSON []byte
		err := rows.Scan(
			&workout.ID,
			&workout.UserID,
			&workout.Username,
			&workout.AvatarURL,
			&workout.Name,
			&workout.Comment,
			&workout.StartedAt,
			&workout.EndedAt,
			&workout.TotalWeight,
			&workout.LikesCount,
			&workout.CommentsCount,
			&workout.UpdatedAt,
			&exercisesJSON,
			&commentsJSON,
			&imagesJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timeline workout: %w", err)
		}
		if len(exercisesJSON) > 0 {
			if err := json.Unmarshal(exercisesJSON, &workout.Exercises); err != nil {
				return nil, fmt.Errorf("failed to unmarshal exercises: %w", err)
			}
		}
		if len(commentsJSON) > 0 {
			if err := json.Unmarshal(commentsJSON, &workout.Comments); err != nil {
				return nil, fmt.Errorf("failed to unmarshal comments: %w", err)
			}
		}
		if len(imagesJSON) > 0 {
			if err := json.Unmarshal(imagesJSON, &workout.Images); err != nil {
				return nil, fmt.Errorf("failed to unmarshal images: %w", err)
			}
		}
		workouts = append(workouts, &workout)
	}
	return workouts, nil
}

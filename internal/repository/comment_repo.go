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

type CommentRepository struct {
	DB *pgxpool.Pool
}

func NewCommentRepository(db *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{
		DB: db,
	}
}

func (r *CommentRepository) CreateComment(
	ctx context.Context,
	userID uuid.UUID,
	workoutID uuid.UUID,
	parentID *uuid.UUID,
	content string,
) (*models.Comment, error) {
	var comment models.Comment

	// userID is passed twice: once as the commenter ($1) and used for access checks
	err := r.DB.QueryRow(ctx, insertCommentQuery, userID, workoutID, parentID, content).Scan(
		&comment.ID,
		&comment.UserID,
		&comment.WorkoutID,
		&comment.ParentID,
		&comment.Content,
		&comment.LikesCount,
		&comment.CreatedAt,
	)
	if err != nil {
		// If no rows returned, the user doesn't have access to this workout
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkoutNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, ErrReferenceViolation
			}
		}
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return &comment, nil
}

func (r *CommentRepository) GetCommentByUserID(
	ctx context.Context,
	id uuid.UUID,
	viewerID uuid.UUID,
) (*models.Comment, error) {
	var comment models.Comment

	err := r.DB.QueryRow(ctx, getCommentByIDQuery, id, viewerID).Scan(
		&comment.ID,
		&comment.UserID,
		&comment.WorkoutID,
		&comment.ParentID,
		&comment.Content,
		&comment.LikesCount,
		&comment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

func (r *CommentRepository) GetCommentsByWorkoutID(
	ctx context.Context,
	workoutID uuid.UUID,
	viewerID uuid.UUID,
	limit int,
	offset int,
) ([]*models.Comment, error) {
	rows, err := r.DB.Query(ctx, getCommentsByWorkoutIDQuery, workoutID, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.WorkoutID,
			&c.ParentID,
			&c.Content,
			&c.LikesCount,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &c)
	}

	return comments, nil
}

func (r *CommentRepository) GetReplies(
	ctx context.Context,
	commentID uuid.UUID,
	viewerID uuid.UUID,
) ([]*models.Comment, error) {
	rows, err := r.DB.Query(ctx, getRepliesByCommentIDQuery, commentID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var c models.Comment
		err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.WorkoutID,
			&c.ParentID,
			&c.Content,
			&c.LikesCount,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}
		comments = append(comments, &c)
	}

	return comments, nil
}

func (r *CommentRepository) DeleteComment(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteCommentByIDQuery, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrCommentNotFound
	}

	return nil
}

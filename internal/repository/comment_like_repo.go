package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotsu1/jimu-backend/internal/models"
)

type CommentLikeRepository struct {
	DB *pgxpool.Pool
}

func NewCommentLikeRepository(db *pgxpool.Pool) *CommentLikeRepository {
	return &CommentLikeRepository{
		DB: db,
	}
}

func (r *CommentLikeRepository) LikeComment(
	ctx context.Context,
	userID uuid.UUID,
	commentID uuid.UUID,
) (*models.CommentLike, error) {
	var like models.CommentLike

	err := r.DB.QueryRow(ctx, insertCommentLikeQuery, userID, commentID).Scan(
		&like.UserID,
		&like.CommentID,
		&like.CreatedAt,
	)
	if err != nil {
		// ON CONFLICT DO NOTHING returns no rows
		if errors.Is(err, pgx.ErrNoRows) {
			// Try to fetch existing like
			existingLike, getErr := r.GetCommentLikeByID(ctx, userID, commentID)
			if getErr != nil {
				if errors.Is(getErr, ErrCommentLikeNotFound) {
					return nil, ErrCommentInteractionNotAllowed
				}
				return nil, getErr
			}
			return existingLike, nil
		}
		return nil, fmt.Errorf("failed to like comment: %w", err)
	}

	return &like, nil
}

func (r *CommentLikeRepository) UnlikeComment(
	ctx context.Context,
	userID uuid.UUID,
	commentID uuid.UUID,
) error {
	commandTag, err := r.DB.Exec(ctx, deleteCommentLikeQuery, userID, commentID)
	if err != nil {
		return fmt.Errorf("failed to unlike comment: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrCommentLikeNotFound
	}

	return nil
}

func (r *CommentLikeRepository) GetCommentLikeByID(
	ctx context.Context,
	userID uuid.UUID,
	commentID uuid.UUID,
) (*models.CommentLike, error) {
	var like models.CommentLike

	err := r.DB.QueryRow(ctx, getCommentLikeQuery, userID, commentID).Scan(
		&like.UserID,
		&like.CommentID,
		&like.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommentLikeNotFound
		}
		return nil, fmt.Errorf("failed to get like: %w", err)
	}

	return &like, nil
}

func (r *CommentLikeRepository) GetCommentLikesByCommentID(
	ctx context.Context,
	commentID uuid.UUID,
	viewerID uuid.UUID,
	limit int,
	offset int,
) ([]*models.CommentLikeDetail, error) {
	rows, err := r.DB.Query(ctx, getCommentLikesByCommentIDQuery, commentID, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comment likes: %w", err)
	}
	defer rows.Close()

	likes := []*models.CommentLikeDetail{}

	for rows.Next() {
		var l models.CommentLikeDetail
		err = rows.Scan(
			&l.UserID,
			&l.CommentID,
			&l.CreatedAt,
			&l.Username,
			&l.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment like detail: %w", err)
		}
		likes = append(likes, &l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return likes, nil
}

func (r *CommentLikeRepository) IsCommentLiked(
	ctx context.Context,
	userID uuid.UUID,
	commentID uuid.UUID,
) (bool, error) {
	var isLiked bool
	err := r.DB.QueryRow(ctx, isCommentLikedQuery, userID, commentID).Scan(&isLiked)
	if err != nil {
		return false, fmt.Errorf("failed to check if liked: %w", err)
	}
	return isLiked, nil
}

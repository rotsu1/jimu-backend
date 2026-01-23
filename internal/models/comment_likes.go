package models

import (
	"time"

	"github.com/google/uuid"
)

type CommentLike struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	CommentID uuid.UUID `json:"comment_id" db:"comment_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

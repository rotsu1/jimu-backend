package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     uuid.UUID  `json:"user_id" db:"user_id"`
	WorkoutID  uuid.UUID  `json:"workout_id" db:"workout_id"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Content    string     `json:"content" db:"content"`
	LikesCount int        `json:"likes_count" db:"likes_count"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

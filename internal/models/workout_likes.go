package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkoutLike struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	WorkoutID uuid.UUID `json:"workout_id" db:"workout_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type WorkoutLikeDetail struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	WorkoutID uuid.UUID `json:"workout_id" db:"workout_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Username  string    `json:"username" db:"username"`
	AvatarURL *string   `json:"avatar_url" db:"avatar_url"`
}

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

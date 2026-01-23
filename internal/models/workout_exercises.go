package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkoutExercise struct {
	ID               uuid.UUID `json:"id" db:"id"`
	WorkoutID        uuid.UUID `json:"workout_id" db:"workout_id"`
	ExerciseID       uuid.UUID `json:"exercise_id" db:"exercise_id"`
	OrderIndex       *int      `json:"order_index,omitempty" db:"order_index"`
	Memo             *string   `json:"memo,omitempty" db:"memo"`
	RestTimerSeconds *int      `json:"rest_timer_seconds,omitempty" db:"rest_timer_seconds"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

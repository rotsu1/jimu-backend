package models

import (
	"time"

	"github.com/google/uuid"
)

type RoutineExercise struct {
	ID               uuid.UUID `json:"id" db:"id"`
	RoutineID        uuid.UUID `json:"routine_id" db:"routine_id"`
	ExerciseID       uuid.UUID `json:"exercise_id" db:"exercise_id"`
	OrderIndex       *int      `json:"order_index,omitempty" db:"order_index"`
	RestTimerSeconds *int      `json:"rest_timer_seconds,omitempty" db:"rest_timer_seconds"`
	Note             *string   `json:"note,omitempty" db:"note"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

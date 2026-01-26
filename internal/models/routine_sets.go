package models

import (
	"time"

	"github.com/google/uuid"
)

type RoutineSet struct {
	ID                uuid.UUID `json:"id" db:"id"`
	RoutineExerciseID uuid.UUID `json:"routine_exercise_id" db:"routine_exercise_id"`
	Weight            *float64  `json:"weight,omitempty" db:"weight"`
	Reps              *int      `json:"reps,omitempty" db:"reps"`
	OrderIndex        *int      `json:"order_index,omitempty" db:"order_index"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateRoutineSetRequest struct {
	Weight     *float64 `json:"weight" db:"weight"`
	Reps       *int     `json:"reps" db:"reps"`
	OrderIndex *int     `json:"order_index" db:"order_index"`
}

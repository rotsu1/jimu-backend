package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkoutSet struct {
	ID                uuid.UUID `json:"id" db:"id"`
	WorkoutExerciseID uuid.UUID `json:"workout_exercise_id" db:"workout_exercise_id"`
	Weight            *float64  `json:"weight,omitempty" db:"weight"`
	Reps              *int      `json:"reps,omitempty" db:"reps"`
	IsCompleted       bool      `json:"is_completed" db:"is_completed"`
	OrderIndex        *int      `json:"order_index,omitempty" db:"order_index"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateWorkoutSetRequest struct {
	Weight      *float64 `json:"weight" db:"weight"`
	Reps        *int     `json:"reps" db:"reps"`
	IsCompleted *bool    `json:"is_completed" db:"is_completed"`
	OrderIndex  *int     `json:"order_index" db:"order_index"`
}

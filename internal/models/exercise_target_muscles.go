package models

import (
	"time"

	"github.com/google/uuid"
)

type ExerciseTargetMuscle struct {
	ExerciseID uuid.UUID `json:"exercise_id" db:"exercise_id"`
	MuscleID   uuid.UUID `json:"muscle_id" db:"muscle_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

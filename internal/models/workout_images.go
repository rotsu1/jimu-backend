package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkoutImage struct {
	ID           uuid.UUID `json:"id" db:"id"`
	WorkoutID    uuid.UUID `json:"workout_id" db:"workout_id"`
	StoragePath  string    `json:"storage_path" db:"storage_path"`
	DisplayOrder *int      `json:"display_order,omitempty" db:"display_order"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

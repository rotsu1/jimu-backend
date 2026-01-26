package models

import (
	"time"

	"github.com/google/uuid"
)

type Workout struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	Name            *string   `json:"name,omitempty" db:"name"`
	Comment         *string   `json:"comment,omitempty" db:"comment"`
	StartedAt       time.Time `json:"started_at" db:"started_at"`
	EndedAt         time.Time `json:"ended_at,omitempty" db:"ended_at"`
	DurationSeconds int       `json:"duration_seconds,omitempty" db:"duration_seconds"`
	TotalWeight     float64   `json:"total_weight,omitempty" db:"total_weight"`
	LikesCount      int       `json:"likes_count" db:"likes_count"`
	CommentsCount   int       `json:"comments_count" db:"comments_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateWorkoutRequest struct {
	Name            *string    `json:"name" db:"name"`
	Comment         *string    `json:"comment" db:"comment"`
	StartedAt       *time.Time `json:"started_at" db:"started_at"`
	EndedAt         *time.Time `json:"ended_at" db:"ended_at"`
	DurationSeconds *int       `json:"duration_seconds" db:"duration_seconds"`
}

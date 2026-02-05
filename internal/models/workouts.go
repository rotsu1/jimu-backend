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

type TimelineWorkout struct {
	ID            uuid.UUID                  `json:"id" db:"id"`
	UserID        uuid.UUID                  `json:"user_id" db:"user_id"`
	Username      string                     `json:"username" db:"username"`
	AvatarURL     *string                    `json:"avatar_url" db:"avatar_url"`
	Name          *string                    `json:"name,omitempty" db:"name"`
	Comment       *string                    `json:"comment,omitempty" db:"comment"`
	StartedAt     time.Time                  `json:"started_at" db:"started_at"`
	EndedAt       time.Time                  `json:"ended_at,omitempty" db:"ended_at"`
	TotalWeight   float64                    `json:"total_weight,omitempty" db:"total_weight"`
	LikesCount    int                        `json:"likes_count" db:"likes_count"`
	CommentsCount int                        `json:"comments_count" db:"comments_count"`
	UpdatedAt     time.Time                  `json:"updated_at" db:"updated_at"`
	Exercises     []*TimelineWorkoutExercise `json:"exercises" db:"exercises"`
	Comments      []*TimelineWorkoutComment  `json:"comments" db:"comments"`
	Images        []*TimelineWorkoutImages   `json:"images" db:"images"`
}

type TimelineWorkoutExercise struct {
	ID         uuid.UUID             `json:"id" db:"id"`
	ExerciseID uuid.UUID             `json:"exercise_id" db:"exercise_id"`
	Name       string                `json:"name" db:"name"`
	OrderIndex int                   `json:"order_index,omitempty" db:"order_index"`
	Sets       []*TimelineWorkoutSet `json:"sets" db:"sets"`
}

type TimelineWorkoutSet struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Weight     *float64  `json:"weight,omitempty" db:"weight"`
	Reps       *int      `json:"reps,omitempty" db:"reps"`
	OrderIndex int       `json:"order_index,omitempty" db:"order_index"`
}

type TimelineWorkoutComment struct {
	ID            uuid.UUID                 `json:"id" db:"id"`
	UserID        uuid.UUID                 `json:"user_id" db:"user_id"`
	Content       string                    `json:"content" db:"content"`
	LikesCount    int                       `json:"likes_count" db:"likes_count"`
	CreatedAt     time.Time                 `json:"created_at" db:"created_at"`
	Username      string                    `json:"username" db:"username"`
	AvatarURL     *string                   `json:"avatar_url" db:"avatar_url"`
	ChildComments []*TimelineWorkoutComment `json:"comments" db:"comments"`
}

type TimelineWorkoutImages struct {
	ID           uuid.UUID `json:"id" db:"id"`
	StoragePath  string    `json:"storage_path" db:"storage_path"`
	DisplayOrder int       `json:"display_order,omitempty" db:"display_order"`
}

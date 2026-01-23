package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID               uuid.UUID `json:"id" db:"id"`
	Username         string    `json:"username" db:"username"`
	DisplayName      string    `json:"display_name" db:"display_name"`
	Bio              string    `json:"bio" db:"bio"`
	Location         string    `json:"location" db:"location"`
	BirthDate        time.Time `json:"birth_date" db:"birth_date"`
	AvatarURL        string    `json:"avatar_url" db:"avatar_url"`
	SubscriptionPlan string    `json:"subscription_plan" db:"subscription_plan"`
	IsPrivateAccount bool      `json:"is_private_account" db:"is_private_account"`
	LastWorkedOutAt  time.Time `json:"last_worked_out_at" db:"last_worked_out_at"`
	TotalWorkouts    int       `json:"total_workouts" db:"total_workouts"`
	CurrentStreak    int       `json:"current_streak" db:"current_streak"`
	TotalWeight      int       `json:"total_weight" db:"total_weight"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateProfileRequest struct {
	Username         *string    `json:"username" db:"username"`
	DisplayName      *string    `json:"display_name" db:"display_name"`
	Bio              *string    `json:"bio" db:"bio"`
	Location         *string    `json:"location" db:"location"`
	BirthDate        *time.Time `json:"birth_date" db:"birth_date"`
	AvatarURL        *string    `json:"avatar_url" db:"avatar_url"`
	SubscriptionPlan *string    `json:"subscription_plan" db:"subscription_plan"`
	IsPrivateAccount *bool      `json:"is_private_account" db:"is_private_account"`
}

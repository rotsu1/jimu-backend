package models

import (
	"time"

	"github.com/google/uuid"
)

type Exercise struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	UserID               uuid.UUID `json:"user_id" db:"user_id"`
	Name                 string    `json:"name" db:"name"`
	SuggestedRestSeconds *int      `json:"suggested_rest_seconds,omitempty" db:"suggested_rest_seconds"`
	Icon                 *string   `json:"icon,omitempty" db:"icon"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type UserIdentity struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	ProviderName   string    `json:"provider_name" db:"provider_name"` // 'google' or 'apple'
	ProviderUserID string    `json:"provider_user_id" db:"provider_user_id"`
	ProviderEmail  *string   `json:"provider_email,omitempty" db:"provider_email"`
	LastSignInAt   time.Time `json:"last_sign_in_at" db:"last_sign_in_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
	ClientIP     *string   `json:"client_ip" db:"client_ip"`
	IsRevoked    bool      `json:"is_revoked" db:"is_revoked"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

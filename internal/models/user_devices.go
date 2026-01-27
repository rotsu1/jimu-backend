package models

import (
	"time"

	"github.com/google/uuid"
)

type UserDevice struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	FCMToken   string    `json:"fcm_token,omitempty" db:"fcm_token"`
	DeviceType string    `json:"device_type,omitempty" db:"device_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

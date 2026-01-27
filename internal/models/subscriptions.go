package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	UserID                uuid.UUID `json:"user_id" db:"user_id"`
	OriginalTransactionID string    `json:"original_transaction_id,omitempty" db:"original_transaction_id"`
	ProductID             string    `json:"product_id,omitempty" db:"product_id"`
	Status                string    `json:"status,omitempty" db:"status"`
	ExpiresAt             time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Environment           string    `json:"environment,omitempty" db:"environment"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

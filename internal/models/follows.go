package models

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	FollowerID  uuid.UUID `json:"follower_id" db:"follower_id"`
	FollowingID uuid.UUID `json:"following_id" db:"following_id"`
	Status      string    `json:"status" db:"status"` // 'pending', 'accepted'
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

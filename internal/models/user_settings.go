package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSetting struct {
	UserID                 uuid.UUID `json:"user_id" db:"user_id"`
	NotifyNewFollower      bool      `json:"notify_new_follower" db:"notify_new_follower"`
	NotifyLikes            bool      `json:"notify_likes" db:"notify_likes"`
	NotifyComments         bool      `json:"notify_comments" db:"notify_comments"`
	SoundEnabled           bool      `json:"sound_enabled" db:"sound_enabled"`
	SoundEffectName        string    `json:"sound_effect_name" db:"sound_effect_name"`
	DefaultTimerSeconds    int       `json:"default_timer_seconds" db:"default_timer_seconds"`
	AutoFillPreviousValues bool      `json:"auto_fill_previous_values" db:"auto_fill_previous_values"`
	UnitWeight             string    `json:"unit_weight" db:"unit_weight"`
	UnitDistance           string    `json:"unit_distance" db:"unit_distance"`
	UnitLength             string    `json:"unit_length" db:"unit_length"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateUserSettingsRequest struct {
	NotifyNewFollower      *bool   `json:"notify_new_follower" db:"notify_new_follower"`
	NotifyLikes            *bool   `json:"notify_likes" db:"notify_likes"`
	NotifyComments         *bool   `json:"notify_comments" db:"notify_comments"`
	SoundEnabled           *bool   `json:"sound_enabled" db:"sound_enabled"`
	SoundEffectName        *string `json:"sound_effect_name" db:"sound_effect_name"`
	DefaultTimerSeconds    *int    `json:"default_timer_seconds" db:"default_timer_seconds"`
	AutoFillPreviousValues *bool   `json:"auto_fill_previous_values" db:"auto_fill_previous_values"`
	UnitWeight             *string `json:"unit_weight" db:"unit_weight"`
	UnitDistance           *string `json:"unit_distance" db:"unit_distance"`
	UnitLength             *string `json:"unit_length" db:"unit_length"`
}

package repository

import "errors"

// Generic errors
var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrReferenceViolation = errors.New("reference violation")
)

// Profile errors
var (
	ErrProfileNotFound       = errors.New("profile not found")
	ErrFailedToUpdateProfile = errors.New("failed to update profile")
	ErrUsernameTaken         = errors.New("username already taken")
)

// Exercise errors
var (
	ErrExerciseNotFound = errors.New("exercise not found")
)

// Workout errors
var (
	ErrWorkoutNotFound = errors.New("workout not found")
)

// WorkoutExercise errors
var (
	ErrWorkoutExerciseNotFound = errors.New("workout exercise not found")
)

// WorkoutSet errors
var (
	ErrWorkoutSetNotFound = errors.New("workout set not found")
)

// WorkoutImage errors
var (
	ErrWorkoutImageNotFound = errors.New("workout image not found")
)

// Routine errors
var (
	ErrRoutineNotFound = errors.New("routine not found")
)

// RoutineExercise errors
var (
	ErrRoutineExerciseNotFound = errors.New("routine exercise not found")
)

// RoutineSet errors
var (
	ErrRoutineSetNotFound = errors.New("routine set not found")
)

// Follow errors
var (
	ErrFollowNotFound = errors.New("follow not found")
)

// BlockedUser errors
var (
	ErrBlockedUserNotFound = errors.New("blocked user not found")
	ErrBlocked             = errors.New("cannot follow: a block exists between these users")
)

// Comment errors
var (
	ErrCommentNotFound              = errors.New("comment not found")
	ErrCommentInteractionNotAllowed = errors.New("comment interaction not allowed")
)

// WorkoutLike errors
var (
	ErrWorkoutLikeNotFound          = errors.New("workout like not found")
	ErrWorkoutInteractionNotAllowed = errors.New("workout interaction not allowed")
)

// CommentLike errors
var (
	ErrCommentLikeNotFound = errors.New("comment like not found")
)

// Subscription errors
var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// UserDevice errors
var (
	ErrUserDeviceNotFound = errors.New("user device not found")
)

// UserSession errors
var (
	ErrUserSessionNotFound = errors.New("user session not found")
)

// Muscle errors
var (
	ErrMuscleNotFound     = errors.New("muscle not found")
	ErrUnauthorizedAction = errors.New("unauthorized action")
)

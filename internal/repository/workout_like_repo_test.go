package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestLikeWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	like, err := likeRepo.LikeWorkout(ctx, userID, workout.ID)
	if err != nil {
		t.Fatalf("Failed to like: %v", err)
	}

	if like.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", like.UserID, userID)
	}
	if like.WorkoutID != workout.ID {
		t.Errorf("WorkoutID mismatch: got %v, want %v", like.WorkoutID, workout.ID)
	}
}

func TestLikeWorkoutIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	// Like twice
	likeRepo.LikeWorkout(ctx, userID, workout.ID)
	like, err := likeRepo.LikeWorkout(ctx, userID, workout.ID)
	if err != nil {
		t.Fatalf("Failed to like (idempotent): %v", err)
	}

	if like.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", like.UserID, userID)
	}

	var count int
	err = db.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM public.workout_likes WHERE user_id = $1 AND workout_id = $2",
		userID,
		workout.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 like, got %d", count)
	}
}

func TestLikeWorkoutNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	_, err := likeRepo.LikeWorkout(ctx, userID, uuid.New())
	if !errors.Is(err, ErrWorkoutInteractionNotAllowed) {
		t.Errorf("Expected ErrWorkoutInteractionNotAllowed, but got %v", err)
	}
}

func TestLikeBlockedUserWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	blockedUserID, _, _ := testutil.InsertProfile(ctx, db, "blockeduser")

	_, err := db.Exec(
		ctx,
		"INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		userID,
		blockedUserID,
	)
	if err != nil {
		t.Fatalf("Failed to insert blocked user: %v", err)
	}

	_, err = likeRepo.LikeWorkout(ctx, userID, blockedUserID)
	if !errors.Is(err, ErrWorkoutInteractionNotAllowed) {
		t.Errorf("Expected ErrWorkoutInteractionNotAllowed, but got %v", err)
	}
}

func TestIsLikedWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	// Not liked initially
	isLiked, _ := likeRepo.IsWorkoutLiked(ctx, userID, workout.ID)
	if isLiked {
		t.Error("Should not be liked initially")
	}

	likeRepo.LikeWorkout(ctx, userID, workout.ID)

	// Liked after liking
	isLiked, _ = likeRepo.IsWorkoutLiked(ctx, userID, workout.ID)
	if !isLiked {
		t.Error("Should be liked after liking")
	}
}

func TestIsLikedWorkoutWorkoutNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	isLiked, err := likeRepo.IsWorkoutLiked(ctx, userID, uuid.New())
	if err != nil {
		t.Fatalf("Failed to check if liked: %v", err)
	}
	if isLiked {
		t.Error("Like should not exist for non-existent workout")
	}
}

func TestIsLikedBlockedUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	blockedUserID, _, _ := testutil.InsertProfile(ctx, db, "blockeduser")

	_, err := db.Exec(
		ctx,
		"INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		userID,
		blockedUserID,
	)
	if err != nil {
		t.Fatalf("Failed to insert blocked user: %v", err)
	}

	isLiked, err := likeRepo.IsWorkoutLiked(ctx, userID, blockedUserID)
	if err != nil {
		t.Fatalf("Failed to check if liked: %v", err)
	}
	if isLiked {
		t.Error("Like should not exist for blocked user")
	}
}

func TestUnlikeWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	likeRepo.LikeWorkout(ctx, userID, workout.ID)

	err := likeRepo.UnlikeWorkout(ctx, userID, workout.ID)
	if err != nil {
		t.Fatalf("Failed to unlike: %v", err)
	}

	isLiked, _ := likeRepo.IsWorkoutLiked(ctx, userID, workout.ID)
	if isLiked {
		t.Error("Should not be liked after unliking")
	}
}

func TestUnlikeWorkoutNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	ctx := context.Background()

	err := likeRepo.UnlikeWorkout(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrWorkoutLikeNotFound) {
		t.Errorf("Expected ErrWorkoutLikeNotFound, but got %v", err)
	}
}

func TestGetLikesByWorkoutID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	likeRepo.LikeWorkout(ctx, userID, workout.ID)

	likes, err := likeRepo.GetWorkoutLikesByWorkoutID(ctx, workout.ID, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 1 {
		t.Errorf("Expected 1 like, got %d", len(likes))
	}
}

func TestGetLikesByWorkoutIDWorkoutNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	likes, err := likeRepo.GetWorkoutLikesByWorkoutID(ctx, uuid.New(), userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 0 {
		t.Errorf("Expected 0 likes, got %d", len(likes))
	}
}

func TestGetLikesByWorkoutIDBlockedUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	blockedUserID, _, _ := testutil.InsertProfile(ctx, db, "blockeduser")

	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	likeRepo.LikeWorkout(ctx, userID, workout.ID)

	_, err := db.Exec(
		ctx,
		"INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		userID,
		blockedUserID,
	)
	if err != nil {
		t.Fatalf("Failed to insert blocked user: %v", err)
	}
	likes, err := likeRepo.GetWorkoutLikesByWorkoutID(ctx, workout.ID, blockedUserID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 0 {
		t.Errorf("Expected 0 likes, got %d", len(likes))
	}
}

func TestLikeDeleteOnCascadeWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewWorkoutLikeRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	_, err := likeRepo.LikeWorkout(ctx, userID, workout.ID)
	if err != nil {
		t.Fatalf("Failed to like: %v", err)
	}

	err = workoutRepo.DeleteWorkout(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete workout: %v", err)
	}

	likes, err := likeRepo.GetWorkoutLikesByWorkoutID(ctx, workout.ID, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 0 {
		t.Errorf("Expected 0 likes, got %d", len(likes))
	}
}

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestLikeComment(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	like, err := likeRepo.LikeComment(ctx, userID, comment.ID)
	if err != nil {
		t.Fatalf("Failed to like: %v", err)
	}

	if like.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", like.UserID, userID)
	}
	if like.CommentID != comment.ID {
		t.Errorf("CommentID mismatch: got %v, want %v", like.CommentID, comment.ID)
	}
}

func TestLikeCommentIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	// Like twice
	likeRepo.LikeComment(ctx, userID, comment.ID)
	like, err := likeRepo.LikeComment(ctx, userID, comment.ID)
	if err != nil {
		t.Fatalf("Failed to like (idempotent): %v", err)
	}

	if like.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", like.UserID, userID)
	}
}

func TestCommentLikeNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	ctx := context.Background()

	_, err := likeRepo.LikeComment(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrCommentInteractionNotAllowed) {
		t.Errorf("Expected ErrCommentInteractionNotAllowed, but got %v", err)
	}
}

func TestCommentLikeBlockedUserWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
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

	_, err = likeRepo.LikeComment(ctx, userID, blockedUserID)
	if !errors.Is(err, ErrCommentInteractionNotAllowed) {
		t.Errorf("Expected ErrCommentInteractionNotAllowed, but got %v", err)
	}
}

func TestIsLikedComment(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	// Not liked initially
	isLiked, _ := likeRepo.IsCommentLiked(ctx, userID, comment.ID)
	if isLiked {
		t.Error("Should not be liked initially")
	}

	likeRepo.LikeComment(ctx, userID, comment.ID)

	// Liked after liking
	isLiked, _ = likeRepo.IsCommentLiked(ctx, userID, comment.ID)
	if !isLiked {
		t.Error("Should be liked after liking")
	}
}

func TestIsLikedCommentBlockedUserWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
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

	isLiked, err := likeRepo.IsCommentLiked(ctx, userID, blockedUserID)
	if err != nil {
		t.Fatalf("Failed to check if liked: %v", err)
	}
	if isLiked {
		t.Error("Should not be liked for blocked user")
	}
}

func TestUnlikeComment(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	likeRepo.LikeComment(ctx, userID, comment.ID)

	err := likeRepo.UnlikeComment(ctx, userID, comment.ID)
	if err != nil {
		t.Fatalf("Failed to unlike: %v", err)
	}

	isLiked, _ := likeRepo.IsCommentLiked(ctx, userID, comment.ID)
	if isLiked {
		t.Error("Should not be liked after unliking")
	}
}

func TestUnlikeCommentNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	ctx := context.Background()

	err := likeRepo.UnlikeComment(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrCommentLikeNotFound) {
		t.Errorf("Expected ErrCommentLikeNotFound, but got %v", err)
	}
}

func TestGetCommentLikesByCommentID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	likeRepo.LikeComment(ctx, userID, comment.ID)

	likes, err := likeRepo.GetCommentLikesByCommentID(ctx, comment.ID, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 1 {
		t.Errorf("Expected 1 like, got %d", len(likes))
	}
}

func TestGetCommentLikesByCommentIDCommentNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	likes, err := likeRepo.GetCommentLikesByCommentID(ctx, uuid.New(), userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 0 {
		t.Errorf("Expected 0 likes, got %d", len(likes))
	}
}

func TestGetCommentLikesByCommentIDBlockedUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	blockedUserID, _, _ := testutil.InsertProfile(ctx, db, "blockeduser")

	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	likeRepo.LikeComment(ctx, userID, comment.ID)

	_, err := db.Exec(
		ctx,
		"INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		userID,
		blockedUserID,
	)
	if err != nil {
		t.Fatalf("Failed to insert blocked user: %v", err)
	}

	likes, err := likeRepo.GetCommentLikesByCommentID(ctx, comment.ID, blockedUserID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 0 {
		t.Errorf("Expected 0 likes, got %d", len(likes))
	}
}

func TestLikeDeleteOnCascadeComment(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	likeRepo := NewCommentLikeRepository(db)
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	_, err := likeRepo.LikeComment(ctx, userID, comment.ID)
	if err != nil {
		t.Fatalf("Failed to like: %v", err)
	}

	err = commentRepo.DeleteComment(ctx, comment.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete comment: %v", err)
	}

	likes, err := likeRepo.GetCommentLikesByCommentID(ctx, comment.ID, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get likes: %v", err)
	}
	if len(likes) != 0 {
		t.Errorf("Expected 0 likes, got %d", len(likes))
	}
}

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateComment(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	content := "Great workout!"

	comment, err := commentRepo.CreateComment(ctx, userID, workout.ID, nil, content)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	if comment.ID == uuid.Nil {
		t.Error("Comment ID should not be nil")
	}
	if comment.Content != content {
		t.Errorf("Content mismatch: got %v, want %v", comment.Content, content)
	}
	if comment.ParentID != nil {
		t.Error("ParentID should be nil for top-level comment")
	}
	if comment.LikesCount != 0 {
		t.Errorf("LikesCount should be 0: got %v", comment.LikesCount)
	}
}

func TestCreateReply(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	parentComment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Parent comment")

	reply, err := commentRepo.CreateComment(ctx, userID, workout.ID, &parentComment.ID, "Reply to parent")
	if err != nil {
		t.Fatalf("Failed to create reply: %v", err)
	}

	if reply.ParentID == nil {
		t.Error("ParentID should not be nil for reply")
	}
	if *reply.ParentID != parentComment.ID {
		t.Errorf("ParentID mismatch: got %v, want %v", *reply.ParentID, parentComment.ID)
	}
}

func TestCreateWorkoutCommentCount(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	workout, err := workoutRepo.GetWorkoutByID(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get workout: %v", err)
	}

	if workout.CommentsCount != 1 {
		t.Errorf("CommentsCount mismatch: got %v, want %v", workout.CommentsCount, 1)
	}
}

func TestGetCommentByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	created, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	comment, err := commentRepo.GetCommentByUserID(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get comment: %v", err)
	}

	if comment.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", comment.ID, created.ID)
	}
}

func TestGetCommentByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	ctx := context.Background()

	viewerID, _, _ := testutil.InsertProfile(ctx, db, "viewer")

	_, err := commentRepo.GetCommentByUserID(ctx, uuid.New(), viewerID)
	if !errors.Is(err, ErrCommentNotFound) {
		t.Errorf("Expected ErrCommentNotFound, but got %v", err)
	}
}

func TestGetCommentsByWorkoutID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Comment 1")
	commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Comment 2")

	comments, err := commentRepo.GetCommentsByWorkoutID(ctx, workout.ID, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}
}

func TestGetReplies(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	parent, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Parent")
	commentRepo.CreateComment(ctx, userID, workout.ID, &parent.ID, "Reply 1")
	commentRepo.CreateComment(ctx, userID, workout.ID, &parent.ID, "Reply 2")

	replies, err := commentRepo.GetReplies(ctx, parent.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get replies: %v", err)
	}

	if len(replies) != 2 {
		t.Errorf("Expected 2 replies, got %d", len(replies))
	}
}

func TestDeleteComment(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	comment, _ := commentRepo.CreateComment(ctx, userID, workout.ID, nil, "To be deleted")
	commentID := comment.ID

	err := commentRepo.DeleteComment(ctx, commentID, userID)
	if err != nil {
		t.Fatalf("Failed to delete comment: %v", err)
	}

	_, err = commentRepo.GetCommentByUserID(ctx, commentID, userID)
	if !errors.Is(err, ErrCommentNotFound) {
		t.Errorf("Expected ErrCommentNotFound, but got %v", err)
	}
}

func TestDeleteCommentNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	ctx := context.Background()

	err := commentRepo.DeleteComment(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrCommentNotFound) {
		t.Errorf("Expected ErrCommentNotFound, but got %v", err)
	}
}

func TestCommentCascadeOnWorkoutDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	commentRepo := NewCommentRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	commentRepo.CreateComment(ctx, userID, workout.ID, nil, "Test comment")

	err := workoutRepo.DeleteWorkout(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete workout: %v", err)
	}

	rows, err := commentRepo.GetCommentsByWorkoutID(ctx, workout.ID, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(rows))
	}
}

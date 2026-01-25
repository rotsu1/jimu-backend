package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestFollow(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	follow, err := repo.Follow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}

	if follow.FollowerID != user1ID {
		t.Errorf("FollowerID mismatch: got %v, want %v", follow.FollowerID, user1ID)
	}
	if follow.FollowingID != user2ID {
		t.Errorf("FollowingID mismatch: got %v, want %v", follow.FollowingID, user2ID)
	}
	if follow.Status != "accepted" {
		t.Errorf("Status mismatch: got %v, want %v", follow.Status, "accepted")
	}
}

func TestFollowIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	// Create two users
	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	// Make the user2 private
	_, err := db.Exec(
		ctx,
		"UPDATE public.profiles SET is_private_account = true WHERE id = $1",
		user2ID,
	)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Follow first time -> Should be 'pending'
	follow1, err := repo.Follow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("First follow failed: %v", err)
	}
	if follow1.Status != "pending" {
		t.Errorf("First follow: expected pending, got %v", follow1.Status)
	}

	// Follow second time (Idempotency check)
	follow2, err := repo.Follow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Second follow failed: %v", err)
	}

	if follow2.Status != "pending" {
		t.Errorf("Second follow: status changed to %v, should still be pending", follow2.Status)
	}
}

func TestFollowBlocked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	_, err := db.Exec(
		ctx,
		"INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		user1ID,
		user2ID,
	)
	if err != nil {
		t.Fatalf("Failed to block users: %v", err)
	}

	_, err = repo.Follow(ctx, user1ID, user2ID)
	if !errors.Is(err, ErrBlocked) {
		t.Errorf("Expected ErrBlocked, but got %v", err)
	}
}

func TestGetFollowStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	_, err := db.Exec(
		ctx,
		"UPDATE public.profiles SET is_private_account = true WHERE id = $1",
		user2ID,
	)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	_, err = repo.Follow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}

	follow, err := repo.GetFollowStatus(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to get follow status: %v", err)
	}

	if follow.Status != "pending" {
		t.Errorf("Status mismatch: got %v, want %v", follow.Status, "pending")
	}
}

func TestGetFollowStatusNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	_, err := repo.GetFollowStatus(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrFollowNotFound) {
		t.Errorf("Expected ErrFollowNotFound, but got %v", err)
	}
}

func TestAcceptFollow(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")
	_, err := db.Exec(
		ctx,
		"UPDATE public.profiles SET is_private_account = true WHERE id = $1",
		user2ID,
	)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	_, err = repo.Follow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}

	err = repo.AcceptFollow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to accept follow: %v", err)
	}

	follow, _ := repo.GetFollowStatus(ctx, user1ID, user2ID)
	if follow.Status != "accepted" {
		t.Errorf("Status should be accepted: got %v", follow.Status)
	}
}

func TestAcceptFollowNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	err := repo.AcceptFollow(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrFollowNotFound) {
		t.Errorf("Expected ErrFollowNotFound, but got %v", err)
	}
}

func TestUnfollow(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	_, err := repo.Follow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}

	err = repo.Unfollow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to unfollow: %v", err)
	}

	_, err = repo.GetFollowStatus(ctx, user1ID, user2ID)
	if !errors.Is(err, ErrFollowNotFound) {
		t.Errorf("Expected ErrFollowNotFound, but got %v", err)
	}
}

func TestUnfollowNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	err := repo.Unfollow(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrFollowNotFound) {
		t.Errorf("Expected ErrFollowNotFound, but got %v", err)
	}
}

func TestGetFollowers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")
	user3ID, _, _ := testutil.InsertProfile(ctx, db, "user3")

	_, err := repo.Follow(ctx, user2ID, user1ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}
	_, err = repo.Follow(ctx, user3ID, user1ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}

	followers, err := repo.GetFollowers(ctx, user1ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get followers: %v", err)
	}

	if len(followers) != 2 {
		t.Errorf("Expected 2 followers, got %d", len(followers))
	}
}

func TestGetFollowing(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewFollowRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")
	user3ID, _, _ := testutil.InsertProfile(ctx, db, "user3")

	_, err := repo.Follow(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}
	_, err = repo.Follow(ctx, user1ID, user3ID)
	if err != nil {
		t.Fatalf("Failed to follow: %v", err)
	}

	following, err := repo.GetFollowing(ctx, user1ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get following: %v", err)
	}

	if len(following) != 2 {
		t.Errorf("Expected 2 following, got %d", len(following))
	}
}

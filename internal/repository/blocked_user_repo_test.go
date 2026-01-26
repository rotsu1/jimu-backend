package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestBlock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	blocked, err := repo.Block(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to block: %v", err)
	}

	if blocked.BlockerID != user1ID {
		t.Errorf("BlockerID mismatch: got %v, want %v", blocked.BlockerID, user1ID)
	}
	if blocked.BlockedID != user2ID {
		t.Errorf("BlockedID mismatch: got %v, want %v", blocked.BlockedID, user2ID)
	}
}

func TestBlockIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	// Block twice
	repo.Block(ctx, user1ID, user2ID)
	blocked, err := repo.Block(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to block (idempotent): %v", err)
	}

	if blocked.BlockerID != user1ID {
		t.Errorf("BlockerID mismatch: got %v, want %v", blocked.BlockerID, user1ID)
	}
}

func TestBlockNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	_, err := repo.Block(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrReferenceViolation) {
		t.Errorf("Expected ErrReferenceViolation, but got %v", err)
	}
}

func TestFollowBlockedUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1") // The Follower
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2") // The Target

	_, err := db.Exec(
		ctx,
		"INSERT INTO public.follows (follower_id, following_id, status) VALUES ($1, $2, 'accepted')",
		user1ID,
		user2ID,
	)
	if err != nil {
		t.Fatalf("Failed to insert follow: %v", err)
	}

	_, err = repo.Block(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to block: %v", err)
	}

	var count int
	err = db.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM public.follows WHERE follower_id = $1 AND following_id = $2",
		user1ID,
		user2ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query follows table: %v", err)
	}
	if count != 0 {
		t.Errorf("Trigger Failure: Follow record still exists after block")
	}

	var followingCount int
	err = db.QueryRow(
		ctx,
		"SELECT following_count FROM public.profiles WHERE id = $1",
		user1ID,
	).Scan(&followingCount)
	if err != nil {
		t.Fatalf("Failed to query user1 profile: %v", err)
	}

	if followingCount != 0 {
		t.Errorf("Counter Failure: User1 following_count should be 0, got %d", followingCount)
	}
}

func TestGetBlockedUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	repo.Block(ctx, user1ID, user2ID)

	blocked, err := repo.GetBlockedUser(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to get blocked user: %v", err)
	}

	if blocked.BlockerID != user1ID {
		t.Errorf("BlockerID mismatch: got %v, want %v", blocked.BlockerID, user1ID)
	}
}

func TestGetBlockedUserNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetBlockedUser(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrBlockedUserNotFound) {
		t.Errorf("Expected ErrBlockedUserNotFound, but got %v", err)
	}
}

func TestIsBlocked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	// Not blocked initially
	isBlocked, _ := repo.IsBlocked(ctx, user1ID, user2ID)
	if isBlocked {
		t.Error("Should not be blocked initially")
	}

	repo.Block(ctx, user1ID, user2ID)

	// Blocked after blocking
	isBlocked, _ = repo.IsBlocked(ctx, user1ID, user2ID)
	if !isBlocked {
		t.Error("Should be blocked after blocking")
	}

	// Also blocked in reverse direction check
	isBlocked, _ = repo.IsBlocked(ctx, user2ID, user1ID)
	if !isBlocked {
		t.Error("Should be blocked in reverse direction check")
	}
}

func TestUnblock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")

	repo.Block(ctx, user1ID, user2ID)

	err := repo.Unblock(ctx, user1ID, user2ID)
	if err != nil {
		t.Fatalf("Failed to unblock: %v", err)
	}

	_, err = repo.GetBlockedUser(ctx, user1ID, user2ID)
	if !errors.Is(err, ErrBlockedUserNotFound) {
		t.Errorf("Expected ErrBlockedUserNotFound, but got %v", err)
	}
}

func TestUnblockNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	err := repo.Unblock(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrBlockedUserNotFound) {
		t.Errorf("Expected ErrBlockedUserNotFound, but got %v", err)
	}
}

func TestGetBlockedUsers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewBlockedUserRepository(db)
	ctx := context.Background()

	user1ID, _, _ := testutil.InsertProfile(ctx, db, "user1")
	user2ID, _, _ := testutil.InsertProfile(ctx, db, "user2")
	user3ID, _, _ := testutil.InsertProfile(ctx, db, "user3")

	repo.Block(ctx, user1ID, user2ID)
	repo.Block(ctx, user1ID, user3ID)

	blockedUsers, err := repo.GetBlockedUsers(ctx, user1ID)
	if err != nil {
		t.Fatalf("Failed to get blocked users: %v", err)
	}

	if len(blockedUsers) != 2 {
		t.Errorf("Expected 2 blocked users, got %d", len(blockedUsers))
	}
}

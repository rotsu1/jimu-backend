package repository

import (
	"context"
	"errors"

	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateUserSession(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	refreshToken := "refresh_token_123"
	userAgent := "TestAgent/1.0"
	clientIP := "127.0.0.1"
	expiresAt := time.Now().Add(24 * time.Hour)

	session, err := repo.CreateSession(ctx, userID, refreshToken, &userAgent, &clientIP, expiresAt)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.ID == uuid.Nil {
		t.Error("Session ID should not be nil")
	}
	if session.RefreshToken != refreshToken {
		t.Errorf("RefreshToken mismatch: got %v, want %v", session.RefreshToken, refreshToken)
	}
	if session.IsRevoked {
		t.Error("Session should not be revoked initially")
	}
}

func TestGetUserSessionByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	created, _ := repo.CreateSession(ctx, userID, "token123", nil, nil, time.Now().Add(time.Hour))

	// Updated: pass userID as viewerID
	session, err := repo.GetSessionByID(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", session.ID, created.ID)
	}
}

func TestGetUserSessionByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	// Updated: pass a random userID as viewerID
	_, err := repo.GetSessionByID(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrUserSessionNotFound) {
		t.Errorf("Expected ErrUserSessionNotFound, but got %v", err)
	}
}

func TestGetUserSessionByRefreshToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	refreshToken := "unique_refresh_token"
	repo.CreateSession(ctx, userID, refreshToken, nil, nil, time.Now().Add(time.Hour))

	session, err := repo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if session.RefreshToken != refreshToken {
		t.Errorf("RefreshToken mismatch: got %v, want %v", session.RefreshToken, refreshToken)
	}
}

func TestGetUserSessionByRefreshTokenExpired(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	// Create expired session
	refreshToken := "expired_token"
	repo.CreateSession(ctx, userID, refreshToken, nil, nil, time.Now().Add(-time.Hour))

	_, err := repo.GetSessionByRefreshToken(ctx, refreshToken)
	if !errors.Is(err, ErrUserSessionNotFound) {
		t.Errorf("Expected ErrUserSessionNotFound for expired token, but got %v", err)
	}
}

func TestGetUserSessionByRefreshTokenRevoked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	refreshToken := "revoked_token"
	session, _ := repo.CreateSession(ctx, userID, refreshToken, nil, nil, time.Now().Add(time.Hour))
	// Updated: pass userID
	repo.RevokeSession(ctx, session.ID, userID)

	_, err := repo.GetSessionByRefreshToken(ctx, refreshToken)
	if !errors.Is(err, ErrUserSessionNotFound) {
		t.Errorf("Expected ErrUserSessionNotFound for revoked token, but got %v", err)
	}
}

func TestGetUserSessionsByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	repo.CreateSession(ctx, userID, "token1", nil, nil, time.Now().Add(time.Hour))
	repo.CreateSession(ctx, userID, "token2", nil, nil, time.Now().Add(time.Hour))

	// Updated: pass userID as viewerID
	sessions, err := repo.GetSessionsByUserID(ctx, userID, userID)
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}

func TestRevokeUserSession(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	session, _ := repo.CreateSession(ctx, userID, "revoke_test", nil, nil, time.Now().Add(time.Hour))

	// Updated: pass userID
	err := repo.RevokeSession(ctx, session.ID, userID)
	if err != nil {
		t.Fatalf("Failed to revoke session: %v", err)
	}

	// Updated: pass userID
	revoked, _ := repo.GetSessionByID(ctx, session.ID, userID)
	if !revoked.IsRevoked {
		t.Error("Session should be revoked")
	}
}

func TestRevokeUserSessionNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	// Updated: pass random userID
	err := repo.RevokeSession(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrUserSessionNotFound) {
		t.Errorf("Expected ErrUserSessionNotFound, but got %v", err)
	}
}

func TestRevokeAllForUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	repo.CreateSession(ctx, userID, "token1", nil, nil, time.Now().Add(time.Hour))
	repo.CreateSession(ctx, userID, "token2", nil, nil, time.Now().Add(time.Hour))

	// Updated: pass userID
	err := repo.RevokeAllSessionsForUser(ctx, userID, userID)
	if err != nil {
		t.Fatalf("Failed to revoke all sessions: %v", err)
	}

	// Updated: pass userID
	sessions, _ := repo.GetSessionsByUserID(ctx, userID, userID)
	for _, s := range sessions {
		if !s.IsRevoked {
			t.Error("All sessions should be revoked")
		}
	}
}

func TestDeleteUserSession(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	session, _ := repo.CreateSession(ctx, userID, "delete_test", nil, nil, time.Now().Add(time.Hour))
	sessionID := session.ID

	// Updated: pass userID
	err := repo.DeleteSession(ctx, sessionID, userID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Updated: pass userID
	_, err = repo.GetSessionByID(ctx, sessionID, userID)
	if !errors.Is(err, ErrUserSessionNotFound) {
		t.Errorf("Expected ErrUserSessionNotFound, but got %v", err)
	}
}

func TestDeleteUserSessionNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserSessionRepository(db)
	ctx := context.Background()

	// Updated: pass random userID
	err := repo.DeleteSession(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrUserSessionNotFound) {
		t.Errorf("Expected ErrUserSessionNotFound, but got %v", err)
	}
}

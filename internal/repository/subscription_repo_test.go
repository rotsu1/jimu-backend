package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestUpsertSubscription(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	transactionID := "txn_123"
	productID := "premium_monthly"
	status := "active"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	environment := "production"

	sub, err := repo.UpsertSubscription(
		ctx,
		userID,
		transactionID,
		productID,
		status,
		expiresAt,
		environment,
	)
	if err != nil {
		t.Fatalf("Failed to upsert subscription: %v", err)
	}

	if sub.ID == uuid.Nil {
		t.Error("Subscription ID should not be nil")
	}
	if sub.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", sub.UserID, userID)
	}
	if sub.Status != status {
		t.Errorf("Status mismatch: got %v, want %v", sub.Status, status)
	}
}

func TestUpsertSubscriptionIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	transactionID := "txn_123"
	productID := "premium_monthly"
	status := "active"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	environment := "production"

	// First upsert
	sub1, _ := repo.UpsertSubscription(
		ctx,
		userID,
		transactionID,
		productID,
		status,
		expiresAt,
		environment,
	)

	// Second upsert with different status
	newStatus := "expired"
	sub2, err := repo.UpsertSubscription(
		ctx,
		userID,
		transactionID,
		productID,
		newStatus,
		expiresAt,
		environment,
	)
	if err != nil {
		t.Fatalf("Failed to upsert subscription: %v", err)
	}

	// Should be same ID
	if sub1.ID != sub2.ID {
		t.Errorf("ID should be the same: got %v, want %v", sub2.ID, sub1.ID)
	}
	// Status should be updated
	if sub2.Status != newStatus {
		t.Errorf("Status should be updated: got %v, want %v", sub2.Status, newStatus)
	}
}

func TestGetSubscriptionByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	transactionID := "txn_456"
	status := "active"
	productID := "premium_monthly"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	environment := "production"
	_, err := repo.UpsertSubscription(
		ctx,
		userID,
		transactionID,
		productID,
		status,
		expiresAt,
		environment,
	)
	if err != nil {
		t.Fatalf("Failed to upsert subscription: %v", err)
	}

	sub, err := repo.GetSubscriptionByUserID(ctx, userID, userID)
	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}

	if sub.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", sub.UserID, userID)
	}
}

func TestGetSubscriptionByUserIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	_, err := repo.GetSubscriptionByUserID(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("Expected ErrSubscriptionNotFound, but got %v", err)
	}
}

func TestGetSubscriptionByTransactionID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	transactionID := "txn_789"
	status := "active"
	productID := "premium_monthly"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	environment := "production"
	_, err := repo.UpsertSubscription(
		ctx,
		userID,
		transactionID,
		productID,
		status,
		expiresAt,
		environment,
	)
	if err != nil {
		t.Fatalf("Failed to upsert subscription: %v", err)
	}

	sub, err := repo.GetSubscriptionByTransactionID(ctx, transactionID, userID)
	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}

	if sub.OriginalTransactionID != transactionID {
		t.Errorf("TransactionID mismatch: got %v, want %v", sub.OriginalTransactionID, transactionID)
	}
}

func TestDeleteSubscription(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	transactionID := "txn_delete"
	status := "active"
	productID := "premium_monthly"
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	environment := "production"
	repo.UpsertSubscription(ctx, userID, transactionID, productID, status, expiresAt, environment)

	err := repo.DeleteSubscription(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to delete subscription: %v", err)
	}

	_, err = repo.GetSubscriptionByUserID(ctx, userID, uuid.New())
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("Expected ErrSubscriptionNotFound, but got %v", err)
	}
}

func TestDeleteSubscriptionNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	err := repo.DeleteSubscription(ctx, uuid.New())
	if !errors.Is(err, ErrSubscriptionNotFound) {
		t.Errorf("Expected ErrSubscriptionNotFound, but got %v", err)
	}
}

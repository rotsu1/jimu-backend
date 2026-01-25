package repository

import (
	"context"
	"testing"
	"time"

	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestNewUserUpsertGoogleUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	googleID := "1234567890"
	email := "test@example.com"

	// Upsert
	profile, err := repo.UpsertGoogleUser(ctx, googleID, email)
	if err != nil {
		t.Fatalf("Failed to upsert: %v", err)
	}

	identity, err := repo.GetIdentityByProvider(ctx, "google", googleID)
	if err != nil {
		t.Fatalf("Identity not found: %v", err)
	}

	// Check if the ID is the same
	if identity.UserID != profile.ID {
		t.Errorf("ID mismatch: got %v, want %v", profile.ID, identity.UserID)
	}

	// Check if the email is the same
	if *identity.ProviderEmail != email {
		t.Errorf("Email mismatch: got %s, want %s", *identity.ProviderEmail, email)
	}

	// Check if the created at is not zero
	if identity.CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}

	// Check if the profile was created
	var exists bool
	err = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)", profile.ID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check profiles table: %v", err)
	}

	if !exists {
		t.Error("Trigger failed: Profile row was not auto-generated")
	}
}

func TestExistingUserUpsertGoogleUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	googleID := "1234567890"
	email := "test@example.com"

	// First upsert
	_, err := repo.UpsertGoogleUser(ctx, googleID, email)
	if err != nil {
		t.Fatalf("Failed to upsert: %v", err)
	}

	identity, err := repo.GetIdentityByProvider(ctx, "google", googleID)
	if err != nil {
		t.Fatalf("Identity not found: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	// Second upsert
	_, err = repo.UpsertGoogleUser(ctx, googleID, email)
	if err != nil {
		t.Fatalf("Failed to upsert: %v", err)
	}

	updatedIdentity, err := repo.GetIdentityByProvider(ctx, "google", googleID)
	if err != nil {
		t.Fatalf("Identity not found: %v", err)
	}

	// Check no duplicate records were created
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM user_identities WHERE provider_user_id = $1", googleID).Scan(&count)
	if count != 1 {
		t.Errorf("Duplicate records created: got %d, want 1", count)
	}

	// check if updated identities data is after the original identity
	if !updatedIdentity.LastSignInAt.After(identity.LastSignInAt) {
		t.Errorf("LastSignInAt was not updated: got %v, want %v", updatedIdentity.LastSignInAt, identity.LastSignInAt)
	}
	if !updatedIdentity.UpdatedAt.After(identity.UpdatedAt) {
		t.Errorf("UpdatedAt was not updated: got %v, want %v", updatedIdentity.UpdatedAt, identity.UpdatedAt)
	}
}

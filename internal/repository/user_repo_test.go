package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestGetIdentityByProvider(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	googleID := "1234567890"
	email := "test@example.com"

	_, err := db.Exec(ctx, insertUserIdentityQuery, googleID, email)
	if err != nil {
		t.Fatalf("Failed to insert identity: %v", err)
	}

	identity, err := repo.GetIdentityByProvider(ctx, "google", googleID)
	if err != nil {
		t.Fatalf("Identity not found: %v", err)
	}

	if identity == nil {
		t.Error("Identity not found")
	}

	if identity.ProviderUserID != googleID {
		t.Errorf("ProviderUserID mismatch: got %s, want %s", identity.ProviderUserID, googleID)
	}
	if identity.ProviderEmail == nil || *identity.ProviderEmail != email {
		t.Errorf("ProviderEmail mismatch: got %v, want %s", identity.ProviderEmail, email)
	}
}

func TestGetNotFoundIdentityByProvider(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, err := repo.GetIdentityByProvider(ctx, "google", "1234567890")
	if err == nil {
		t.Error("Expected error when identity is not found, but got nil")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("Expected pgx.ErrNoRows, but got %v", err)
	}
}

func TestNullFieldsIdentityByProvider(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	googleID := "1234567890"

	_, err := db.Exec(
		ctx,
		"INSERT INTO user_identities (provider_name, provider_user_id) VALUES ($1, $2)",
		"google",
		googleID,
	)
	if err != nil {
		t.Fatalf("Failed to insert identity: %v", err)
	}

	identity, err := repo.GetIdentityByProvider(ctx, "google", googleID)
	if err != nil {
		t.Fatalf("Identity not found: %v", err)
	}

	if identity.ProviderEmail != nil {
		t.Error("ProviderEmail is not nil")
	}
}

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

// func TestDeleteProfile(t *testing.T) {
// 	db := testutil.SetupTestDB(t)
// 	defer db.Close()
// 	repo := NewUserRepository(db)
// 	ctx := context.Background()

// 	googleID := "1234567890"
// 	email := "test@example.com"

// 	profile, err := repo.UpsertGoogleUser(ctx, googleID, email)
// 	if err != nil {
// 		t.Fatalf("Failed to upsert: %v", err)
// 	}

// }

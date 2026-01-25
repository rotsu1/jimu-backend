package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
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
	var profileExists bool
	err = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM profiles WHERE id = $1)", profile.ID).Scan(&profileExists)
	if err != nil {
		t.Fatalf("Failed to check profiles table: %v", err)
	}

	if !profileExists {
		t.Error("Trigger failed: Profile row was not auto-generated")
	}

	// Check if user settings was created
	var userSettingsExists bool
	err = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM user_settings WHERE user_id = $1)", profile.ID).Scan(&userSettingsExists)
	if err != nil {
		t.Fatalf("Failed to check user_settings table: %v", err)
	}

	if !userSettingsExists {
		t.Error("Trigger failed: User settings row was not auto-generated")
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

func TestExistingGetProfileByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	username := "testuser"

	var id uuid.UUID
	err := db.QueryRow(
		ctx,
		"INSERT INTO profiles (username) VALUES ($1) RETURNING id",
		username,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	profile, err := repo.GetProfileByID(ctx, id, id)
	if err != nil {
		t.Fatalf("Profile not found: %v", err)
	}

	if profile.ID != id {
		t.Errorf("ID mismatch: got %v, want %v", profile.ID, id)
	}
}

func TestNotFoundGetProfileByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	id := uuid.New()
	_, err := repo.GetProfileByID(ctx, id, id)
	if err == nil {
		t.Error("Expected error when profile is not found, but got nil")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("Expected pgx.ErrNoRows, but got %v", err)
	}
}

func TestNullFieldsGetProfileByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	var id uuid.UUID
	username := "testuser"
	err := db.QueryRow(
		ctx,
		"INSERT INTO profiles (username) VALUES ($1) RETURNING id",
		username,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	profile, err := repo.GetProfileByID(ctx, id, id)
	if err != nil {
		t.Fatalf("Profile not found: %v", err)
	}

	if profile.Username != username {
		t.Errorf("Username mismatch: got %v, want %v", profile.Username, username)
	}
	if profile.DisplayName != nil {
		t.Errorf("DisplayName is not nil: got %v, want nil", profile.DisplayName)
	}
	if profile.Bio != nil {
		t.Errorf("Bio is not nil: got %v, want nil", profile.Bio)
	}
	if profile.Location != nil {
		t.Errorf("Location is not nil: got %v, want nil", profile.Location)
	}
	if profile.BirthDate != nil {
		t.Errorf("BirthDate is not nil: got %v, want nil", profile.BirthDate)
	}
	if profile.AvatarURL != nil {
		t.Errorf("AvatarURL is not nil: got %v, want nil", profile.AvatarURL)
	}
	if profile.SubscriptionPlan != nil {
		t.Errorf("SubscriptionPlan is not nil: got %v, want nil", profile.SubscriptionPlan)
	}
	if profile.IsPrivateAccount {
		t.Errorf("IsPrivateAccount is not false: got %v, want false", profile.IsPrivateAccount)
	}
	if profile.LastWorkedOutAt != nil {
		t.Errorf("LastWorkedOutAt is not nil: got %v, want nil", profile.LastWorkedOutAt)
	}
	if profile.TotalWorkouts != 0 {
		t.Errorf("TotalWorkouts is not 0: got %v, want 0", profile.TotalWorkouts)
	}
	if profile.CurrentStreak != 0 {
		t.Errorf("CurrentStreak is not 0: got %v, want 0", profile.CurrentStreak)
	}
	if profile.TotalWeight != 0 {
		t.Errorf("TotalWeight is not 0: got %v, want 0", profile.TotalWeight)
	}
	if profile.CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}
	if profile.UpdatedAt.IsZero() {
		t.Error("UpdatedAt was not set")
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

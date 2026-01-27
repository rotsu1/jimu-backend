package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rotsu1/jimu-backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

// Test User Identity Get Functionality
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

// Test New User Upsert Google User Functionality
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

// Test Existing User Get Profile By ID Functionality
func TestExistingGetProfileByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	id, _, err := testutil.InsertProfile(ctx, db, "testuser")
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

	username := "testuser"
	id, _, err := testutil.InsertProfile(ctx, db, username)

	profile, err := repo.GetProfileByID(ctx, id, id)
	if err != nil {
		t.Fatalf("Profile not found: %v", err)
	}

	if *profile.Username != username {
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

// Test Update Profile Functionality
func TestUpdateProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	id, updatedAt, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	newUsername := "newtestuser"
	displayName := "New Test User"
	updates := models.UpdateProfileRequest{
		Username:    &newUsername,
		DisplayName: &displayName,
	}
	err = repo.UpdateProfile(ctx, id, updates)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	profile, err := repo.GetProfileByID(ctx, id, id)
	if err != nil {
		t.Fatalf("Profile not found: %v", err)
	}

	if *profile.Username != newUsername {
		t.Errorf("Username was not updated: got %v, want %v", profile.Username, newUsername)
	}
	if *profile.DisplayName != displayName {
		t.Errorf("DisplayName was not updated: got %v, want %v", profile.DisplayName, displayName)
	}
	if !profile.UpdatedAt.After(updatedAt) {
		t.Errorf("UpdatedAt was not updated: got %v, want %v", profile.UpdatedAt, updatedAt)
	}
}

func TestNotFoundUpdateProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	id := uuid.New()
	err := repo.UpdateProfile(ctx, id, models.UpdateProfileRequest{})
	if err != nil {
		t.Errorf("Expected to silently fail, but got %v", err)
	}

	displayName := "New Name"
	reqWithData := models.UpdateProfileRequest{
		DisplayName: &displayName,
	}
	err = repo.UpdateProfile(ctx, id, reqWithData)
	if err == nil {
		t.Error("Expected error when profile is not found, but got nil")
	}

	if !errors.Is(err, ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, but got %v", err)
	}
}

func TestUpdateProfileWithNullFields(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	var id uuid.UUID
	username := "testuser"
	displayName := "Test User"
	var updatedAt time.Time
	err := db.QueryRow(
		ctx,
		"INSERT INTO profiles (username, display_name) VALUES ($1, $2) RETURNING id, updated_at",
		username,
		displayName,
	).Scan(&id, &updatedAt)
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	var updatedReq models.UpdateProfileRequest
	displayNameNil := ""
	birthDateNil := time.Time{}
	updatedReq.DisplayName = &displayNameNil
	updatedReq.BirthDate = &birthDateNil

	err = repo.UpdateProfile(ctx, id, updatedReq)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	profile, err := repo.GetProfileByID(ctx, id, id)
	if err != nil {
		t.Fatalf("Profile not found: %v", err)
	}

	if profile.DisplayName != nil {
		t.Errorf("DisplayName is not nil: got %v, want nil", profile.DisplayName)
	}
	if !profile.UpdatedAt.After(updatedAt) {
		t.Errorf("UpdatedAt was not updated: got %v, want %v", profile.UpdatedAt, updatedAt)
	}
}

func TestNothingToUpdateUpdateProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	id, updatedAt, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	err = repo.UpdateProfile(ctx, id, models.UpdateProfileRequest{})
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	profile, err := repo.GetProfileByID(ctx, id, id)
	if err != nil {
		t.Fatalf("Profile not found: %v", err)
	}

	if profile.UpdatedAt != updatedAt {
		t.Errorf("UpdatedAt was not the same: got %v, want %v", profile.UpdatedAt, updatedAt)
	}
}

func TestConstraintViolationUpdateProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	username1 := "testuser1"
	id1, _, err := testutil.InsertProfile(ctx, db, username1)
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}
	username2 := "testuser2"
	_, _, err = testutil.InsertProfile(ctx, db, username2)
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}
	updatedReq := models.UpdateProfileRequest{
		Username: &username2,
	}
	err = repo.UpdateProfile(ctx, id1, updatedReq)

	if err == nil {
		t.Error("Expected primary key violation error, but got nil")
	}
	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("Expected ErrUsernameTaken, but got %v", err)
	}
}

// Test Delete Profile Functionality
func TestDeleteProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	googleID := "1234567890"
	email := "test@example.com"

	profile, err := repo.UpsertGoogleUser(ctx, googleID, email)
	if err != nil {
		t.Fatalf("Failed to upsert: %v", err)
	}

	id := profile.ID

	err = repo.DeleteProfile(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete profile: %v", err)
	}

	_, err = repo.GetProfileByID(ctx, id, id)
	if err == nil {
		t.Error("Expected profile to be deleted, but got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("Expected pgx.ErrNoRows, but got %v", err)
	}

	_, err = repo.GetUserSettingsByID(ctx, id)
	if err == nil {
		t.Error("Expected user settings to be deleted, but got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("Expected pgx.ErrNoRows, but got %v", err)
	}

	_, err = repo.GetIdentityByProvider(ctx, "google", googleID)
	if err == nil {
		t.Error("Expected identity to be deleted, but got nil")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("Expected pgx.ErrNoRows, but got %v", err)
	}
}

func TestNotFoundDeleteProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	id := uuid.New()
	err := repo.DeleteProfile(ctx, id)
	if err == nil {
		t.Error("Expected error when profile is not found, but got nil")
	}
	if !errors.Is(err, ErrProfileNotFound) {
		t.Errorf("Expected ErrProfileNotFound, but got %v", err)
	}
}

// Test GetIdentitiesByUserID Functionality
func TestGetIdentitiesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	// 1. Create a user
	googleID := "google-123"
	email := "test@example.com"
	profile, err := repo.UpsertGoogleUser(ctx, googleID, email)
	if err != nil {
		t.Fatalf("Failed to upsert user: %v", err)
	}

	// 2. Fetch identities
	identities, err := repo.GetIdentitiesByUserID(ctx, profile.ID)
	if err != nil {
		t.Fatalf("Failed to get identities by user ID: %v", err)
	}

	// 3. Assertions
	if len(identities) != 1 {
		t.Errorf("Expected 1 identity, got %d", len(identities))
	}
	if identities[0].ProviderName != "google" {
		t.Errorf("Expected provider 'google', got %s", identities[0].ProviderName)
	}
	if identities[0].ProviderUserID != googleID {
		t.Errorf("Expected provider user ID %s, got %s", googleID, identities[0].ProviderUserID)
	}
	if identities[0].UserID != profile.ID {
		t.Errorf("Expected user ID %v, got %v", profile.ID, identities[0].UserID)
	}

	// 4. Manually verify DB count
	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM user_identities WHERE user_id = $1", profile.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count identities: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1 from DB, got %d", count)
	}
}

func TestGetIdentitiesByUserID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	randomID := uuid.New()
	identities, err := repo.GetIdentitiesByUserID(ctx, randomID)
	if err != nil {
		t.Fatalf("Failed to get identities for non-existent user: %v", err)
	}
	if len(identities) != 0 {
		t.Errorf("Expected 0 identities, got %d", len(identities))
	}
}

// Test DeleteIdentity Functionality
func TestDeleteIdentity(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	ctx := context.Background()

	// 1. Create user with an identity
	googleID := "google-delete-test"
	email := "delete@example.com"
	profile, err := repo.UpsertGoogleUser(ctx, googleID, email)
	if err != nil {
		t.Fatalf("Failed to upsert user: %v", err)
	}

	// 2. Verify identity exists
	identities, err := repo.GetIdentitiesByUserID(ctx, profile.ID)
	if err != nil || len(identities) != 1 {
		t.Fatalf("Pre-condition failed: identity should exist")
	}

	// 3. Delete the identity
	err = repo.DeleteIdentity(ctx, profile.ID, "google", googleID)
	if err != nil {
		t.Fatalf("Failed to delete identity: %v", err)
	}

	// 4. Verify identity is gone
	identities, err = repo.GetIdentitiesByUserID(ctx, profile.ID)
	if err != nil {
		t.Fatalf("Failed to get identities after deletion: %v", err)
	}
	if len(identities) != 0 {
		t.Errorf("Expected 0 identities after deletion, got %d", len(identities))
	}

	// Double check via direct query
	_, err = repo.GetIdentityByProvider(ctx, "google", googleID)
	if err == nil {
		t.Error("Expected error (NoRows) when fetching deleted identity, got nil")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("Expected pgx.ErrNoRows, got %v", err)
	}
}

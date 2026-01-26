package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateRoutine(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Push Day"

	routine, err := repo.CreateRoutine(ctx, userID, name)
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}

	if routine.ID == uuid.Nil {
		t.Error("Routine ID should not be nil")
	}
	if routine.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", routine.UserID, userID)
	}
	if routine.Name != name {
		t.Errorf("Name mismatch: got %v, want %v", routine.Name, name)
	}
	if routine.CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}
}

func TestGetRoutineByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	created, err := repo.CreateRoutine(ctx, userID, "Test Routine")
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}

	// Updated: pass userID as viewerID
	routine, err := repo.GetRoutineByID(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get routine: %v", err)
	}

	if routine.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", routine.ID, created.ID)
	}
}

func TestGetRoutineByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	// Start with a valid user for the viewer
	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	_, err := repo.GetRoutineByID(ctx, uuid.New(), userID)
	if err == nil {
		t.Error("Expected error when routine is not found, but got nil")
	}

	if !errors.Is(err, ErrRoutineNotFound) {
		t.Errorf("Expected ErrRoutineNotFound, but got %v", err)
	}
}

func TestGetRoutineBlocked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}
	blockedUserID, _, err := testutil.InsertProfile(ctx, db, "blockeduser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = db.Exec(
		ctx,
		"INSERT INTO blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		userID,
		blockedUserID,
	)
	if err != nil {
		t.Fatalf("Failed to insert blocked user: %v", err)
	}

	blocked, err := repo.CreateRoutine(ctx, blockedUserID, "Blocked Routine")
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}

	_, err = repo.GetRoutineByID(ctx, blocked.ID, userID)
	if err == nil {
		t.Error("Expected error when routine is blocked, but got nil")
	}

	if !errors.Is(err, ErrRoutineNotFound) {
		t.Errorf("Expected ErrRoutineNotFound, but got %v", err)
	}
}

func TestGetRoutinesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = repo.CreateRoutine(ctx, userID, "Push Day")
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}
	_, err = repo.CreateRoutine(ctx, userID, "Pull Day")
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}

	// Updated: pass userID as viewerID
	routines, err := repo.GetRoutinesByUserID(ctx, userID, userID)
	if err != nil {
		t.Fatalf("Failed to get routines: %v", err)
	}

	if len(routines) != 2 {
		t.Errorf("Expected 2 routines, got %d", len(routines))
	}

	// Should be ordered by name
	if routines[0].Name != "Pull Day" {
		t.Errorf("Expected first routine to be 'Pull Day', got %s", routines[0].Name)
	}
}

func TestUpdateRoutine(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	routine, err := repo.CreateRoutine(ctx, userID, "Original")
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}

	newName := "Updated Routine"
	// Updated: pass userID as owner
	err = repo.UpdateRoutine(ctx, routine.ID, models.UpdateRoutineRequest{
		Name: &newName,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to update routine: %v", err)
	}

	updated, err := repo.GetRoutineByID(ctx, routine.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get routine: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("Name was not updated: got %v, want %v", updated.Name, newName)
	}
	if !updated.UpdatedAt.After(routine.UpdatedAt) {
		t.Errorf("UpdatedAt was not updated")
	}
}

func TestUpdateRoutineNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	// Create a user ID for the owner check
	userID := uuid.New()

	newName := "Test"
	// Updated: pass userID
	err := repo.UpdateRoutine(ctx, uuid.New(), models.UpdateRoutineRequest{
		Name: &newName,
	}, userID)
	if err == nil {
		t.Error("Expected error when routine is not found, but got nil")
	}

	if !errors.Is(err, ErrRoutineNotFound) {
		t.Errorf("Expected ErrRoutineNotFound, but got %v", err)
	}
}

func TestUpdateRoutineNoChange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	routine, err := repo.CreateRoutine(ctx, userID, "Original")
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}

	// Updated: pass userID
	err = repo.UpdateRoutine(ctx, routine.ID, models.UpdateRoutineRequest{}, userID)
	if err != nil {
		t.Fatalf("Expected no error for empty update, got: %v", err)
	}

	updated, err := repo.GetRoutineByID(ctx, routine.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get routine: %v", err)
	}

	if updated.UpdatedAt != routine.UpdatedAt {
		t.Errorf("UpdatedAt should not have changed")
	}
}

func TestDeleteRoutine(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	routine, err := repo.CreateRoutine(ctx, userID, "Test")
	if err != nil {
		t.Fatalf("Failed to create routine: %v", err)
	}

	routineID := routine.ID

	// Updated: pass userID for owner check
	err = repo.DeleteRoutine(ctx, routineID, userID)
	if err != nil {
		t.Fatalf("Failed to delete routine: %v", err)
	}

	_, err = repo.GetRoutineByID(ctx, routineID, userID)
	if err == nil {
		t.Error("Expected routine to be deleted, but got nil")
	}
	if !errors.Is(err, ErrRoutineNotFound) {
		t.Errorf("Expected ErrRoutineNotFound, but got %v", err)
	}
}

func TestDeleteRoutineNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewRoutineRepository(db)
	ctx := context.Background()

	userID := uuid.New()

	err := repo.DeleteRoutine(ctx, uuid.New(), userID)
	if err == nil {
		t.Error("Expected error when routine is not found, but got nil")
	}

	if !errors.Is(err, ErrRoutineNotFound) {
		t.Errorf("Expected ErrRoutineNotFound, but got %v", err)
	}
}

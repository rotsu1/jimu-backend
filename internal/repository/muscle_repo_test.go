package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestGetAllMuscles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	_, err := testutil.InsertMuscle(ctx, db, "TestMuscle")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	muscles, err := repo.GetAllMuscles(ctx)
	if err != nil {
		t.Fatalf("Failed to get muscles: %v", err)
	}

	if len(muscles) == 0 {
		t.Error("Expected at least one muscle")
	}
}

func TestGetMuscleByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	insertedMuscle, err := testutil.InsertMuscle(ctx, db, "TestMuscle")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	muscle, err := repo.GetMuscleByID(ctx, insertedMuscle.ID)
	if err != nil {
		t.Fatalf("Failed to get muscle: %v", err)
	}

	if muscle.ID != insertedMuscle.ID {
		t.Errorf("ID mismatch: got %v, want %v", muscle.ID, insertedMuscle.ID)
	}
}

func TestGetMuscleByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	_, err := repo.GetMuscleByID(ctx, uuid.New())
	if !errors.Is(err, ErrMuscleNotFound) {
		t.Errorf("Expected ErrMuscleNotFound, but got %v", err)
	}
}

func TestGetMuscleByName(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	insertedMuscle, err := testutil.InsertMuscle(ctx, db, "TestMuscle")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	muscle, err := repo.GetMuscleByName(ctx, insertedMuscle.Name)
	if err != nil {
		t.Fatalf("Failed to get muscle: %v", err)
	}

	if muscle.Name != insertedMuscle.Name {
		t.Errorf("Name mismatch: got %v, want %v", muscle.Name, insertedMuscle.Name)
	}
}

func TestGetMuscleByNameNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	_, err := repo.GetMuscleByName(ctx, "NonexistentMuscle")
	if !errors.Is(err, ErrMuscleNotFound) {
		t.Errorf("Expected ErrMuscleNotFound, but got %v", err)
	}
}

func TestCreateMuscle(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	err := testutil.InsertSysAdmin(ctx, db, userID)
	if err != nil {
		t.Fatalf("Failed to insert sys admin: %v", err)
	}

	_, err = repo.CreateMuscle(ctx, "TestMuscle", userID)
	if err != nil {
		t.Fatalf("Failed to create muscle: %v", err)
	}
}

func TestCreateMuscleAlreadyExists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	err := testutil.InsertSysAdmin(ctx, db, userID)
	if err != nil {
		t.Fatalf("Failed to insert sys admin: %v", err)
	}

	_, err = repo.CreateMuscle(ctx, "TestMuscle", userID)
	if err != nil {
		t.Fatalf("Failed to create muscle: %v", err)
	}

	_, err = repo.CreateMuscle(ctx, "TestMuscle", userID)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("Expected ErrAlreadyExists, but got %v", err)
	}
}

func TestCreateMuscleNotAuthorized(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	_, err := repo.CreateMuscle(ctx, "TestMuscle", userID)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !errors.Is(err, ErrUnauthorizedAction) {
		t.Errorf("Expected ErrUnauthorizedAction, but got %v", err)
	}
}

func TestDeleteMuscle(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	err := testutil.InsertSysAdmin(ctx, db, userID)
	if err != nil {
		t.Fatalf("Failed to insert sys admin: %v", err)
	}

	muscle, err := repo.CreateMuscle(ctx, "TestMuscle", userID)
	if err != nil {
		t.Fatalf("Failed to create muscle: %v", err)
	}

	err = repo.DeleteMuscle(ctx, muscle.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete muscle: %v", err)
	}
}

func TestDeleteMuscleNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	err := testutil.InsertSysAdmin(ctx, db, userID)
	if err != nil {
		t.Fatalf("Failed to insert sys admin: %v", err)
	}

	err = repo.DeleteMuscle(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrUnauthorizedAction) {
		t.Errorf("Expected ErrUnauthorizedAction, but got %v", err)
	}
}

func TestDeleteMuscleNotAuthorized(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewMuscleRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	err := repo.DeleteMuscle(ctx, uuid.New(), userID)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if !errors.Is(err, ErrUnauthorizedAction) {
		t.Errorf("Expected ErrUnauthorizedAction, but got %v", err)
	}
}

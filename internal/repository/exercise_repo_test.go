package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Bench Press"
	restSeconds := 90
	icon := "dumbbell"

	exercise, err := repo.CreateExercise(ctx, &userID, name, &restSeconds, &icon, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	if exercise.ID == uuid.Nil {
		t.Error("Exercise ID should not be nil")
	}
	if *exercise.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", exercise.UserID, userID)
	}
	if exercise.Name != name {
		t.Errorf("Name mismatch: got %v, want %v", exercise.Name, name)
	}
	if *exercise.SuggestedRestSeconds != restSeconds {
		t.Errorf("SuggestedRestSeconds mismatch: got %v, want %v", *exercise.SuggestedRestSeconds, restSeconds)
	}
	if *exercise.Icon != icon {
		t.Errorf("Icon mismatch: got %v, want %v", *exercise.Icon, icon)
	}
	if exercise.CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}
}

func TestCreateExerciseWithNullOptionalFields(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	exercise, err := repo.CreateExercise(ctx, &userID, "Squat", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	if exercise.SuggestedRestSeconds != nil {
		t.Errorf("SuggestedRestSeconds should be nil: got %v", exercise.SuggestedRestSeconds)
	}
	if exercise.Icon != nil {
		t.Errorf("Icon should be nil: got %v", exercise.Icon)
	}
}

func TestGetExerciseByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Deadlift"
	created, err := repo.CreateExercise(ctx, &userID, name, nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	exercise, err := repo.GetExerciseByID(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get exercise: %v", err)
	}

	if exercise.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", exercise.ID, created.ID)
	}
	if exercise.Name != name {
		t.Errorf("Name mismatch: got %v, want %v", exercise.Name, name)
	}
}

func TestGetExerciseByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = repo.GetExerciseByID(ctx, uuid.New(), userID)
	if err == nil {
		t.Error("Expected error when exercise is not found, but got nil")
	}

	if !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("Expected ErrExerciseNotFound, but got %v", err)
	}
}

func TestGetExerciseByIDBlocked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	blockedID, _, err := testutil.InsertProfile(ctx, db, "blockeduser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = db.Exec(
		ctx,
		`INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)`,
		userID,
		blockedID,
	)
	if err != nil {
		t.Fatalf("Failed to insert block: %v", err)
	}

	exercise, err := repo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	_, err = repo.GetExerciseByID(ctx, exercise.ID, blockedID)
	if err == nil {
		t.Error("Expected error when exercise is not found, but got nil")
	}

	if !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("Expected ErrExerciseNotFound, but got %v", err)
	}
}

func TestGetExercisesByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = repo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}
	_, err = repo.CreateExercise(ctx, &userID, "Squat", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	exercises, err := repo.GetExercisesByUserID(ctx, userID, userID)
	if err != nil {
		t.Fatalf("Failed to get exercises: %v", err)
	}

	if len(exercises) != 2 {
		t.Errorf("Expected 2 exercises, got %d", len(exercises))
	}

	// Should be ordered by name
	if exercises[0].Name != "Bench Press" {
		t.Errorf("Expected first exercise to be 'Bench Press', got %s", exercises[0].Name)
	}
}

func TestGetExercisesByUserIDBlocked(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	blockedID, _, err := testutil.InsertProfile(ctx, db, "blockeduser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = db.Exec(
		ctx,
		`INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)`,
		userID,
		blockedID,
	)
	if err != nil {
		t.Fatalf("Failed to insert block: %v", err)
	}

	_, err = repo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	exercises, err := repo.GetExercisesByUserID(ctx, userID, blockedID)
	if err != nil {
		t.Fatalf("Failed to get exercises: %v", err)
	}

	if len(exercises) != 0 {
		t.Errorf("Expected 0 exercises, got %d", len(exercises))
	}
}

func TestUpdateExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	exercise, err := repo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	newName := "Incline Bench Press"
	newRest := 120
	err = repo.UpdateExercise(ctx, exercise.ID, models.UpdateExerciseRequest{
		Name:                 &newName,
		SuggestedRestSeconds: &newRest,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to update exercise: %v", err)
	}

	updated, err := repo.GetExerciseByID(ctx, exercise.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get exercise: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("Name was not updated: got %v, want %v", updated.Name, newName)
	}
	if *updated.SuggestedRestSeconds != newRest {
		t.Errorf("SuggestedRestSeconds was not updated: got %v, want %v", *updated.SuggestedRestSeconds, newRest)
	}
	if !updated.UpdatedAt.After(exercise.UpdatedAt) {
		t.Errorf("UpdatedAt was not updated")
	}
}

func TestUpdateExerciseNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	newName := "Test"
	err = repo.UpdateExercise(ctx, uuid.New(), models.UpdateExerciseRequest{
		Name: &newName,
	}, userID)
	if err == nil {
		t.Error("Expected error when exercise is not found, but got nil")
	}

	if !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("Expected ErrExerciseNotFound, but got %v", err)
	}
}

func TestUpdateExerciseNoChange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	exercise, err := repo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	err = repo.UpdateExercise(ctx, exercise.ID, models.UpdateExerciseRequest{}, userID)
	if err != nil {
		t.Fatalf("Expected no error for empty update, got: %v", err)
	}

	updated, err := repo.GetExerciseByID(ctx, exercise.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get exercise: %v", err)
	}

	if updated.UpdatedAt != exercise.UpdatedAt {
		t.Errorf("UpdatedAt should not have changed: got %v, want %v", updated.UpdatedAt, exercise.UpdatedAt)
	}
}

func TestUpdateExerciseZeroToNull(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	rest := 90
	icon := "dumbbell"
	exercise, err := repo.CreateExercise(ctx, &userID, "Bench Press", &rest, &icon, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	// Set to zero/empty to trigger NULL conversion
	zeroRest := 0
	emptyIcon := ""
	err = repo.UpdateExercise(ctx, exercise.ID, models.UpdateExerciseRequest{
		SuggestedRestSeconds: &zeroRest,
		Icon:                 &emptyIcon,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to update exercise: %v", err)
	}

	updated, err := repo.GetExerciseByID(ctx, exercise.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get exercise: %v", err)
	}

	if updated.SuggestedRestSeconds != nil {
		t.Errorf("SuggestedRestSeconds should be nil: got %v", updated.SuggestedRestSeconds)
	}
	if updated.Icon != nil {
		t.Errorf("Icon should be nil: got %v", updated.Icon)
	}
}

func TestDeleteExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	exercise, err := repo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	exerciseID := exercise.ID

	err = repo.DeleteExercise(ctx, exerciseID, userID)
	if err != nil {
		t.Fatalf("Failed to delete exercise: %v", err)
	}

	_, err = repo.GetExerciseByID(ctx, exerciseID, userID)
	if err == nil {
		t.Error("Expected exercise to be deleted, but got nil")
	}
	if !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("Expected ErrExerciseNotFound, but got %v", err)
	}
}

func TestDeleteExerciseNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	err = repo.DeleteExercise(ctx, uuid.New(), userID)
	if err == nil {
		t.Error("Expected error when exercise is not found, but got nil")
	}

	if !errors.Is(err, ErrExerciseNotFound) {
		t.Errorf("Expected ErrExerciseNotFound, but got %v", err)
	}
}

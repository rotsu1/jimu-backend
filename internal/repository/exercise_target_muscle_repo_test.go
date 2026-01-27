package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestAddExerciseTargetMuscle(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	etmRepo := NewExerciseTargetMuscleRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)

	muscle, err := testutil.InsertMuscle(ctx, db, "Bench Press")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	etm, err := etmRepo.AddTargetMuscle(ctx, exercise.ID, muscle.ID, userID)
	if err != nil {
		t.Fatalf("Failed to add exercise target muscle: %v", err)
	}

	if etm.ExerciseID != exercise.ID {
		t.Errorf("ExerciseID mismatch: got %v, want %v", etm.ExerciseID, exercise.ID)
	}
	if etm.MuscleID != muscle.ID {
		t.Errorf("MuscleID mismatch: got %v, want %v", etm.MuscleID, muscle.ID)
	}
}

func TestAddExerciseTargetMuscleIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	etmRepo := NewExerciseTargetMuscleRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Squat", nil, nil, userID)

	muscle, err := testutil.InsertMuscle(ctx, db, "Squat")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	// Add twice
	etmRepo.AddTargetMuscle(ctx, exercise.ID, muscle.ID, userID)
	etm, err := etmRepo.AddTargetMuscle(ctx, exercise.ID, muscle.ID, userID)
	if err != nil {
		t.Fatalf("Failed to add exercise target muscle (idempotent): %v", err)
	}

	if etm.ExerciseID != exercise.ID {
		t.Errorf("ExerciseID mismatch: got %v, want %v", etm.ExerciseID, exercise.ID)
	}
}

func TestAddExerciseTargetMuscleExerciseNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	etmRepo := NewExerciseTargetMuscleRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Deadlift", nil, nil, userID)

	_, err := etmRepo.AddTargetMuscle(ctx, exercise.ID, uuid.New(), userID)
	if !errors.Is(err, ErrReferenceViolation) {
		t.Errorf("Expected ErrReferenceViolation, but got %v", err)
	}
}

func TestGetExerciseTargetMusclesByExerciseID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	etmRepo := NewExerciseTargetMuscleRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Deadlift", nil, nil, userID)

	muscle1, err := testutil.InsertMuscle(ctx, db, "Back")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}
	muscle2, err := testutil.InsertMuscle(ctx, db, "Biceps")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	_, err = etmRepo.AddTargetMuscle(ctx, exercise.ID, muscle1.ID, userID)
	if err != nil {
		t.Fatalf("Failed to add exercise target muscle: %v", err)
	}
	_, err = etmRepo.AddTargetMuscle(ctx, exercise.ID, muscle2.ID, userID)
	if err != nil {
		t.Fatalf("Failed to add exercise target muscle: %v", err)
	}

	targetMuscles, err := etmRepo.GetByExerciseID(ctx, exercise.ID)
	if err != nil {
		t.Fatalf("Failed to get exercise target muscles: %v", err)
	}

	if len(targetMuscles) != 2 {
		t.Errorf("Expected 2 target muscles, got %d", len(targetMuscles))
	}
}

func TestSetExerciseTargetMuscles(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	etmRepo := NewExerciseTargetMuscleRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Row", nil, nil, userID)

	muscle1, err := testutil.InsertMuscle(ctx, db, "Back")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}
	muscle2, err := testutil.InsertMuscle(ctx, db, "Biceps")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}
	muscle3, err := testutil.InsertMuscle(ctx, db, "Triceps")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	// First set 2 muscles
	err = etmRepo.SetTargetMuscles(ctx, exercise.ID, []uuid.UUID{muscle1.ID, muscle2.ID}, userID)
	if err != nil {
		t.Fatalf("Failed to set exercise target muscles: %v", err)
	}

	targetMuscles, _ := etmRepo.GetByExerciseID(ctx, exercise.ID)
	if len(targetMuscles) != 2 {
		t.Errorf("Expected 2 target muscles, got %d", len(targetMuscles))
	}

	// Then set 1 muscle (should replace)
	err = etmRepo.SetTargetMuscles(ctx, exercise.ID, []uuid.UUID{muscle3.ID}, userID)
	if err != nil {
		t.Fatalf("Failed to set exercise target muscles: %v", err)
	}

	targetMuscles, _ = etmRepo.GetByExerciseID(ctx, exercise.ID)
	if len(targetMuscles) != 1 {
		t.Errorf("Expected 1 target muscle after replace, got %d", len(targetMuscles))
	}
	if targetMuscles[0].MuscleID != muscle3.ID {
		t.Errorf("MuscleID should be muscle3.ID")
	}
}

func TestRemoveExerciseTargetMuscle(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	etmRepo := NewExerciseTargetMuscleRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Press", nil, nil, userID)

	muscle, err := testutil.InsertMuscle(ctx, db, "Press")
	if err != nil {
		t.Fatalf("Failed to insert muscle: %v", err)
	}

	etmRepo.AddTargetMuscle(ctx, exercise.ID, muscle.ID, userID)

	err = etmRepo.RemoveTargetMuscle(ctx, exercise.ID, muscle.ID, userID)
	if err != nil {
		t.Fatalf("Failed to remove exercise target muscle: %v", err)
	}

	targetMuscles, _ := etmRepo.GetByExerciseID(ctx, exercise.ID)
	if len(targetMuscles) != 0 {
		t.Errorf("Expected 0 target muscles after removal, got %d", len(targetMuscles))
	}
}

func TestRemoveExerciseTargetMuscleNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	etmRepo := NewExerciseTargetMuscleRepository(db)
	ctx := context.Background()

	err := etmRepo.RemoveTargetMuscle(ctx, uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected ErrNotFound, but got %v", err)
	}
}

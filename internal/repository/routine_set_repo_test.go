package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateRoutineSet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Push Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Bench Press", nil, nil)
	// Updated: pass userID
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)

	weight := 100.0
	reps := 10
	orderIndex := 1

	// Updated: pass userID
	rs, err := rsRepo.CreateRoutineSet(ctx, re.ID, &weight, &reps, &orderIndex, userID)
	if err != nil {
		t.Fatalf("Failed to create routine set: %v", err)
	}

	if rs.ID == uuid.Nil {
		t.Error("RoutineSet ID should not be nil")
	}
	if *rs.Weight != weight {
		t.Errorf("Weight mismatch: got %v, want %v", *rs.Weight, weight)
	}
	if *rs.Reps != reps {
		t.Errorf("Reps mismatch: got %v, want %v", *rs.Reps, reps)
	}
}

func TestGetRoutineSetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Pull Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Row", nil, nil)
	// Updated: pass userID
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)

	// Updated: pass userID
	created, _ := rsRepo.CreateRoutineSet(ctx, re.ID, nil, nil, nil, userID)

	rs, err := rsRepo.GetRoutineSetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get routine set: %v", err)
	}

	if rs.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", rs.ID, created.ID)
	}
}

func TestGetRoutineSetByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	ctx := context.Background()

	_, err := rsRepo.GetRoutineSetByID(ctx, uuid.New())
	if !errors.Is(err, ErrRoutineSetNotFound) {
		t.Errorf("Expected ErrRoutineSetNotFound, but got %v", err)
	}
}

func TestGetRoutineSetsByRoutineExerciseID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Leg Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Squat", nil, nil)
	// Updated: pass userID
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)

	order1 := 2
	order2 := 1
	// Updated: pass userID
	rsRepo.CreateRoutineSet(ctx, re.ID, nil, nil, &order1, userID)
	rsRepo.CreateRoutineSet(ctx, re.ID, nil, nil, &order2, userID)

	sets, err := rsRepo.GetRoutineSetsByRoutineExerciseID(ctx, re.ID)
	if err != nil {
		t.Fatalf("Failed to get routine sets: %v", err)
	}

	if len(sets) != 2 {
		t.Errorf("Expected 2 sets, got %d", len(sets))
	}

	// Should be ordered by order_index ASC
	if *sets[0].OrderIndex != 1 {
		t.Errorf("Expected first set to have order_index 1")
	}
}

func TestUpdateRoutineSet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Arm Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Curl", nil, nil)
	// Updated: pass userID
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)

	// Updated: pass userID
	rs, _ := rsRepo.CreateRoutineSet(ctx, re.ID, nil, nil, nil, userID)

	newWeight := 25.0
	newReps := 12
	// Updated: pass userID
	err := rsRepo.UpdateRoutineSet(ctx, rs.ID, models.UpdateRoutineSetRequest{
		Weight: &newWeight,
		Reps:   &newReps,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to update routine set: %v", err)
	}

	updated, _ := rsRepo.GetRoutineSetByID(ctx, rs.ID)
	if *updated.Weight != newWeight {
		t.Errorf("Weight was not updated: got %v, want %v", *updated.Weight, newWeight)
	}
	if *updated.Reps != newReps {
		t.Errorf("Reps was not updated: got %v, want %v", *updated.Reps, newReps)
	}
}

func TestUpdateRoutineSetNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	weight := 100.0
	// Updated: pass userID
	err := rsRepo.UpdateRoutineSet(ctx, uuid.New(), models.UpdateRoutineSetRequest{Weight: &weight}, userID)
	if !errors.Is(err, ErrRoutineSetNotFound) {
		t.Errorf("Expected ErrRoutineSetNotFound, but got %v", err)
	}
}

func TestDeleteRoutineSet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Full Body")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Deadlift", nil, nil)
	// Updated: pass userID
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)

	// Updated: pass userID
	rs, _ := rsRepo.CreateRoutineSet(ctx, re.ID, nil, nil, nil, userID)
	rsID := rs.ID

	// Updated: pass userID
	err := rsRepo.DeleteRoutineSet(ctx, rsID, userID)
	if err != nil {
		t.Fatalf("Failed to delete routine set: %v", err)
	}

	_, err = rsRepo.GetRoutineSetByID(ctx, rsID)
	if !errors.Is(err, ErrRoutineSetNotFound) {
		t.Errorf("Expected ErrRoutineSetNotFound, but got %v", err)
	}
}

func TestDeleteRoutineSetNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	// Updated: pass userID
	err := rsRepo.DeleteRoutineSet(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrRoutineSetNotFound) {
		t.Errorf("Expected ErrRoutineSetNotFound, but got %v", err)
	}
}

func TestDeleteRoutineSetOnDeleteRoutineExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	rsRepo := NewRoutineSetRepository(db)
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Leg Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Squat", nil, nil)
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)

	rs, _ := rsRepo.CreateRoutineSet(ctx, re.ID, nil, nil, nil, userID)
	rsID := rs.ID

	err := reRepo.DeleteRoutineExercise(ctx, re.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete routine exercise: %v", err)
	}

	_, err = rsRepo.GetRoutineSetByID(ctx, rsID)
	if !errors.Is(err, ErrRoutineSetNotFound) {
		t.Errorf("Expected ErrRoutineSetNotFound, but got %v", err)
	}
}

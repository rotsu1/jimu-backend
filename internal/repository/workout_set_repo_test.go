package repository

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateWorkoutSet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.Create(ctx, userID, "Bench Press", nil, nil)
	we, _ := weRepo.Create(ctx, workout.ID, exercise.ID, nil, nil, nil)

	weight := 100.0
	reps := 10
	orderIndex := 1

	ws, err := wsRepo.Create(ctx, we.ID, &weight, &reps, false, &orderIndex, userID)
	if err != nil {
		t.Fatalf("Failed to create workout set: %v", err)
	}

	if ws.ID == uuid.Nil {
		t.Error("WorkoutSet ID should not be nil")
	}
	if *ws.Weight != weight {
		t.Errorf("Weight mismatch: got %v, want %v", *ws.Weight, weight)
	}
	if *ws.Reps != reps {
		t.Errorf("Reps mismatch: got %v, want %v", *ws.Reps, reps)
	}
	if ws.IsCompleted {
		t.Error("IsCompleted should be false")
	}
}

func TestCreateWorkoutSetNotFoundWorkoutExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	_, err := wsRepo.Create(ctx, uuid.New(), nil, nil, false, nil, userID)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}
	if !errors.Is(err, ErrReferenceViolation) {
		t.Errorf("Expected ErrReferenceViolation, but got %v", err)
	}
}

func TestCreateWorkoutSetSyncWorkoutStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	weight := 80.0
	reps := 20
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.Create(ctx, userID, "Bench Press", nil, nil)
	we, err := weRepo.Create(ctx, workout.ID, exercise.ID, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create workout exercise: %v", err)
	}
	_, err = wsRepo.Create(ctx, we.ID, &weight, &reps, false, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create workout set: %v", err)
	}

	workout, err = workoutRepo.GetWorkoutByID(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get workout: %v", err)
	}
	if math.Abs(workout.TotalWeight-weight*float64(reps)) > 0.0001 {
		t.Errorf("Total weight mismatch: got %v, want %v", workout.TotalWeight, weight*float64(reps))
	}
}

func TestGetWorkoutSetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.Create(ctx, userID, "Squat", nil, nil)
	we, _ := weRepo.Create(ctx, workout.ID, exercise.ID, nil, nil, nil)

	created, _ := wsRepo.Create(ctx, we.ID, nil, nil, false, nil, userID)

	ws, err := wsRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get workout set: %v", err)
	}

	if ws.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", ws.ID, created.ID)
	}
}

func TestGetWorkoutSetByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	ctx := context.Background()

	_, err := wsRepo.GetByID(ctx, uuid.New())
	if !errors.Is(err, ErrWorkoutSetNotFound) {
		t.Errorf("Expected ErrWorkoutSetNotFound, but got %v", err)
	}
}

func TestGetWorkoutSetsByWorkoutExerciseID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.Create(ctx, userID, "Deadlift", nil, nil)
	we, _ := weRepo.Create(ctx, workout.ID, exercise.ID, nil, nil, nil)

	order1 := 2
	order2 := 1
	wsRepo.Create(ctx, we.ID, nil, nil, false, &order1, userID)
	wsRepo.Create(ctx, we.ID, nil, nil, true, &order2, userID)

	sets, err := wsRepo.GetByWorkoutExerciseID(ctx, we.ID)
	if err != nil {
		t.Fatalf("Failed to get workout sets: %v", err)
	}

	if len(sets) != 2 {
		t.Errorf("Expected 2 sets, got %d", len(sets))
	}

	// Should be ordered by order_index ASC
	if *sets[0].OrderIndex != 1 {
		t.Errorf("Expected first set to have order_index 1")
	}
}

func TestUpdateWorkoutSet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.Create(ctx, userID, "OHP", nil, nil)
	we, _ := weRepo.Create(ctx, workout.ID, exercise.ID, nil, nil, nil)

	ws, _ := wsRepo.Create(ctx, we.ID, nil, nil, false, nil, userID)

	newWeight := 50.0
	newReps := 8
	isCompleted := true
	err := wsRepo.Update(ctx, ws.ID, userID, models.UpdateWorkoutSetRequest{
		Weight:      &newWeight,
		Reps:        &newReps,
		IsCompleted: &isCompleted,
	})
	if err != nil {
		t.Fatalf("Failed to update workout set: %v", err)
	}

	updated, _ := wsRepo.GetByID(ctx, ws.ID)
	if *updated.Weight != newWeight {
		t.Errorf("Weight was not updated: got %v, want %v", *updated.Weight, newWeight)
	}
	if *updated.Reps != newReps {
		t.Errorf("Reps was not updated: got %v, want %v", *updated.Reps, newReps)
	}
	if !updated.IsCompleted {
		t.Error("IsCompleted was not updated to true")
	}
}

func TestUpdateWorkoutSetNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	ctx := context.Background()

	weight := 100.0
	err := wsRepo.Update(ctx, uuid.New(), uuid.New(), models.UpdateWorkoutSetRequest{Weight: &weight})
	if !errors.Is(err, ErrWorkoutSetNotFound) {
		t.Errorf("Expected ErrWorkoutSetNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutSet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.Create(ctx, userID, "Row", nil, nil)
	we, _ := weRepo.Create(ctx, workout.ID, exercise.ID, nil, nil, nil)

	ws, _ := wsRepo.Create(ctx, we.ID, nil, nil, false, nil, userID)
	wsID := ws.ID

	err := wsRepo.Delete(ctx, wsID, userID)
	if err != nil {
		t.Fatalf("Failed to delete workout set: %v", err)
	}

	_, err = wsRepo.GetByID(ctx, wsID)
	if !errors.Is(err, ErrWorkoutSetNotFound) {
		t.Errorf("Expected ErrWorkoutSetNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutSetNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wsRepo := NewWorkoutSetRepository(db)
	ctx := context.Background()

	err := wsRepo.Delete(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, ErrWorkoutSetNotFound) {
		t.Errorf("Expected ErrWorkoutSetNotFound, but got %v", err)
	}
}

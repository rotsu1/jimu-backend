package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateWorkoutExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	workout, err := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	exercise, err := exerciseRepo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create exercise: %v", err)
	}

	orderIndex := 1
	memo := "Focus on form"
	restTimer := 90

	we, err := weRepo.CreateWorkoutExercise(ctx, workout.ID, exercise.ID, &orderIndex, &memo, &restTimer, userID)
	if err != nil {
		t.Fatalf("Failed to create workout exercise: %v", err)
	}

	if we.ID == uuid.Nil {
		t.Error("WorkoutExercise ID should not be nil")
	}
	if we.WorkoutID != workout.ID {
		t.Errorf("WorkoutID mismatch: got %v, want %v", we.WorkoutID, workout.ID)
	}
	if *we.OrderIndex != orderIndex {
		t.Errorf("OrderIndex mismatch: got %v, want %v", *we.OrderIndex, orderIndex)
	}
}

func TestCreateWorkoutExerciseNotFoundWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	_, err := weRepo.CreateWorkoutExercise(ctx, uuid.New(), exercise.ID, nil, nil, nil, userID)
	if !errors.Is(err, ErrReferenceViolation) {
		t.Errorf("Expected ErrReferenceViolation, but got %v", err)
	}
}

func TestGetWorkoutExerciseByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Squat", nil, nil, userID)

	created, err := weRepo.CreateWorkoutExercise(ctx, workout.ID, exercise.ID, nil, nil, nil, userID)
	if err != nil {
		t.Fatalf("Failed to create workout exercise: %v", err)
	}

	we, err := weRepo.GetWorkoutExerciseByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get workout exercise: %v", err)
	}

	if we.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", we.ID, created.ID)
	}
}

func TestGetWorkoutExerciseByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	ctx := context.Background()

	_, err := weRepo.GetWorkoutExerciseByID(ctx, uuid.New())
	if !errors.Is(err, ErrWorkoutExerciseNotFound) {
		t.Errorf("Expected ErrWorkoutExerciseNotFound, but got %v", err)
	}
}

func TestGetWorkoutExercisesByWorkoutID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise1, _ := exerciseRepo.CreateExercise(ctx, &userID, "Bench Press", nil, nil, userID)
	exercise2, _ := exerciseRepo.CreateExercise(ctx, &userID, "Squat", nil, nil, userID)

	order1 := 2
	order2 := 1
	weRepo.CreateWorkoutExercise(ctx, workout.ID, exercise1.ID, &order1, nil, nil, userID)
	weRepo.CreateWorkoutExercise(ctx, workout.ID, exercise2.ID, &order2, nil, nil, userID)

	exercises, err := weRepo.GetWorkoutExercisesByWorkoutID(ctx, workout.ID)
	if err != nil {
		t.Fatalf("Failed to get workout exercises: %v", err)
	}

	if len(exercises) != 2 {
		t.Errorf("Expected 2 exercises, got %d", len(exercises))
	}

	// Should be ordered by order_index ASC
	if exercises[0].ExerciseID != exercise2.ID {
		t.Errorf("Expected first exercise to be exercise2 (order 1)")
	}
}

func TestUpdateWorkoutExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Deadlift", nil, nil, userID)

	we, _ := weRepo.CreateWorkoutExercise(ctx, workout.ID, exercise.ID, nil, nil, nil, userID)

	newMemo := "Updated memo"
	newOrder := 5
	err := weRepo.UpdateWorkoutExercise(ctx, we.ID, models.UpdateWorkoutExerciseRequest{
		Memo:       &newMemo,
		OrderIndex: &newOrder,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to update workout exercise: %v", err)
	}

	updated, _ := weRepo.GetWorkoutExerciseByID(ctx, we.ID)
	if *updated.Memo != newMemo {
		t.Errorf("Memo was not updated: got %v, want %v", *updated.Memo, newMemo)
	}
	if *updated.OrderIndex != newOrder {
		t.Errorf("OrderIndex was not updated: got %v, want %v", *updated.OrderIndex, newOrder)
	}
}

func TestUpdateWorkoutExerciseNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	newMemo := "Test"
	err := weRepo.UpdateWorkoutExercise(ctx, uuid.New(), models.UpdateWorkoutExerciseRequest{Memo: &newMemo}, userID)
	if !errors.Is(err, ErrWorkoutExerciseNotFound) {
		t.Errorf("Expected ErrWorkoutExerciseNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Press", nil, nil, userID)

	we, _ := weRepo.CreateWorkoutExercise(ctx, workout.ID, exercise.ID, nil, nil, nil, userID)
	weID := we.ID

	err := weRepo.DeleteWorkoutExercise(ctx, weID, userID)
	if err != nil {
		t.Fatalf("Failed to delete workout exercise: %v", err)
	}

	_, err = weRepo.GetWorkoutExerciseByID(ctx, weID)
	if !errors.Is(err, ErrWorkoutExerciseNotFound) {
		t.Errorf("Expected ErrWorkoutExerciseNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutExerciseNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	err := weRepo.DeleteWorkoutExercise(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrWorkoutExerciseNotFound) {
		t.Errorf("Expected ErrWorkoutExerciseNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutExerciseOnDeleteWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	weRepo := NewWorkoutExerciseRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	exercise, _ := exerciseRepo.CreateExercise(ctx, &userID, "Press", nil, nil, userID)
	we, _ := weRepo.CreateWorkoutExercise(ctx, workout.ID, exercise.ID, nil, nil, nil, userID)

	err := workoutRepo.DeleteWorkout(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete workout: %v", err)
	}

	_, err = weRepo.GetWorkoutExerciseByID(ctx, we.ID)
	if !errors.Is(err, ErrWorkoutExerciseNotFound) {
		t.Errorf("Expected ErrWorkoutExerciseNotFound, but got %v", err)
	}
}

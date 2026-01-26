package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateRoutineExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Push Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Bench Press", nil, nil)

	orderIndex := 1
	restTimer := 90
	memo := "Focus on chest"

	// Updated: pass userID
	re, err := reRepo.CreateRoutineExercise(
		ctx,
		routine.ID,
		exercise.ID,
		&orderIndex,
		&restTimer,
		&memo,
		userID,
	)
	if err != nil {
		t.Fatalf("Failed to create routine exercise: %v", err)
	}

	if re.ID == uuid.Nil {
		t.Error("RoutineExercise ID should not be nil")
	}
	if *re.OrderIndex != orderIndex {
		t.Errorf("OrderIndex mismatch: got %v, want %v", *re.OrderIndex, orderIndex)
	}
	if *re.Memo != memo {
		t.Errorf("Memo mismatch: got %v, want %v", *re.Memo, memo)
	}
}

func TestGetRoutineExerciseByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Pull Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Row", nil, nil)

	// Updated: pass userID
	created, _ := reRepo.CreateRoutineExercise(
		ctx,
		routine.ID,
		exercise.ID,
		nil,
		nil,
		nil,
		userID,
	)

	re, err := reRepo.GetRoutineExerciseByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get routine exercise: %v", err)
	}

	if re.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", re.ID, created.ID)
	}
}

func TestGetRoutineExerciseByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	ctx := context.Background()

	_, err := reRepo.GetRoutineExerciseByID(ctx, uuid.New())
	if !errors.Is(err, ErrRoutineExerciseNotFound) {
		t.Errorf("Expected ErrRoutineExerciseNotFound, but got %v", err)
	}
}

func TestGetRoutineExercisesByRoutineID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Leg Day")
	exercise1, _ := exerciseRepo.Create(ctx, userID, "Squat", nil, nil)
	exercise2, _ := exerciseRepo.Create(ctx, userID, "Leg Press", nil, nil)

	order1 := 2
	order2 := 1
	// Updated: pass userID
	reRepo.CreateRoutineExercise(ctx, routine.ID, exercise1.ID, &order1, nil, nil, userID)
	reRepo.CreateRoutineExercise(ctx, routine.ID, exercise2.ID, &order2, nil, nil, userID)

	exercises, err := reRepo.GetRoutineExercisesByRoutineID(ctx, routine.ID)
	if err != nil {
		t.Fatalf("Failed to get routine exercises: %v", err)
	}

	if len(exercises) != 2 {
		t.Errorf("Expected 2 exercises, got %d", len(exercises))
	}

	// Should be ordered by order_index ASC
	if exercises[0].ExerciseID != exercise2.ID {
		t.Errorf("Expected first exercise to be exercise2 (order 1)")
	}
}

func TestUpdateRoutineExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Chest Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Incline Press", nil, nil)

	// Updated: pass userID
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)

	newMemo := "Updated memo"
	newOrder := 5
	// Updated: pass userID
	err := reRepo.UpdateRoutineExercise(ctx, re.ID, models.UpdateRoutineExerciseRequest{
		Memo:       &newMemo,
		OrderIndex: &newOrder,
	}, userID)
	if err != nil {
		t.Fatalf("Failed to update routine exercise: %v", err)
	}

	updated, _ := reRepo.GetRoutineExerciseByID(ctx, re.ID)
	if *updated.Memo != newMemo {
		t.Errorf("Memo was not updated: got %v, want %v", *updated.Memo, newMemo)
	}
	if *updated.OrderIndex != newOrder {
		t.Errorf("OrderIndex was not updated: got %v, want %v", *updated.OrderIndex, newOrder)
	}
}

func TestUpdateRoutineExerciseNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	newMemo := "Test"
	// Updated: pass userID
	err := reRepo.UpdateRoutineExercise(
		ctx,
		uuid.New(),
		models.UpdateRoutineExerciseRequest{Memo: &newMemo},
		userID,
	)
	if !errors.Is(err, ErrRoutineExerciseNotFound) {
		t.Errorf("Expected ErrRoutineExerciseNotFound, but got %v", err)
	}
}

func TestDeleteRoutineExercise(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	routineRepo := NewRoutineRepository(db)
	exerciseRepo := NewExerciseRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	routine, _ := routineRepo.CreateRoutine(ctx, userID, "Back Day")
	exercise, _ := exerciseRepo.Create(ctx, userID, "Pullup", nil, nil)

	// Updated: pass userID
	re, _ := reRepo.CreateRoutineExercise(ctx, routine.ID, exercise.ID, nil, nil, nil, userID)
	reID := re.ID

	// Updated: pass userID
	err := reRepo.DeleteRoutineExercise(ctx, reID, userID)
	if err != nil {
		t.Fatalf("Failed to delete routine exercise: %v", err)
	}

	_, err = reRepo.GetRoutineExerciseByID(ctx, reID)
	if !errors.Is(err, ErrRoutineExerciseNotFound) {
		t.Errorf("Expected ErrRoutineExerciseNotFound, but got %v", err)
	}
}

func TestDeleteRoutineExerciseNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	reRepo := NewRoutineExerciseRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	// Updated: pass userID
	err := reRepo.DeleteRoutineExercise(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrRoutineExerciseNotFound) {
		t.Errorf("Expected ErrRoutineExerciseNotFound, but got %v", err)
	}
}

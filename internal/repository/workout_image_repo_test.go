package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/repository/testutil"
)

func TestCreateWorkoutImage(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	storagePath := "/images/workout123.jpg"
	displayOrder := 1

	wi, err := wiRepo.CreateWorkoutImage(ctx, workout.ID, storagePath, displayOrder, userID)
	if err != nil {
		t.Fatalf("Failed to create workout image: %v", err)
	}

	if wi.ID == uuid.Nil {
		t.Error("WorkoutImage ID should not be nil")
	}
	if wi.StoragePath != storagePath {
		t.Errorf("StoragePath mismatch: got %v, want %v", wi.StoragePath, storagePath)
	}
	if wi.DisplayOrder != displayOrder {
		t.Errorf("DisplayOrder mismatch: got %v, want %v", wi.DisplayOrder, displayOrder)
	}
}

func TestCreateWorkoutImageNotFoundWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	_, err := wiRepo.CreateWorkoutImage(ctx, uuid.New(), "/path/to/image.jpg", 0, userID)
	if !errors.Is(err, ErrReferenceViolation) {
		t.Errorf("Expected ErrReferenceViolation, but got %v", err)
	}
}

func TestGetWorkoutImageByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	created, _ := wiRepo.CreateWorkoutImage(ctx, workout.ID, "/path/to/image.jpg", 0, userID)

	wi, err := wiRepo.GetWorkoutImageByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get workout image: %v", err)
	}

	if wi.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", wi.ID, created.ID)
	}
}

func TestGetWorkoutImageByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	ctx := context.Background()

	_, err := wiRepo.GetWorkoutImageByID(ctx, uuid.New())
	if !errors.Is(err, ErrWorkoutImageNotFound) {
		t.Errorf("Expected ErrWorkoutImageNotFound, but got %v", err)
	}
}

func TestGetWorkoutImagesByWorkoutID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	order1 := 2
	order2 := 1
	wiRepo.CreateWorkoutImage(ctx, workout.ID, "/path/to/image1.jpg", order1, userID)
	wiRepo.CreateWorkoutImage(ctx, workout.ID, "/path/to/image2.jpg", order2, userID)

	images, err := wiRepo.GetWorkoutImagesByWorkoutID(ctx, workout.ID)
	if err != nil {
		t.Fatalf("Failed to get workout images: %v", err)
	}

	if len(images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(images))
	}

	// Should be ordered by display_order ASC
	if images[0].DisplayOrder != 1 {
		t.Errorf("Expected first image to have display_order 1")
	}
}

func TestGetWorkoutImagesByWorkoutIDNotFoundWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	ctx := context.Background()

	rows, err := wiRepo.GetWorkoutImagesByWorkoutID(ctx, uuid.New())
	if err != nil {
		t.Fatalf("Failed to get workout images: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("Expected 0 images, got %d", len(rows))
	}
}

func TestDeleteWorkoutImage(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)

	wi, _ := wiRepo.CreateWorkoutImage(ctx, workout.ID, "/path/to/delete.jpg", 0, userID)
	wiID := wi.ID

	err := wiRepo.DeleteWorkoutImage(ctx, wiID, userID)
	if err != nil {
		t.Fatalf("Failed to delete workout image: %v", err)
	}

	_, err = wiRepo.GetWorkoutImageByID(ctx, wiID)
	if !errors.Is(err, ErrWorkoutImageNotFound) {
		t.Errorf("Expected ErrWorkoutImageNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutImageNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")

	err := wiRepo.DeleteWorkoutImage(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrWorkoutImageNotFound) {
		t.Errorf("Expected ErrWorkoutImageNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutImageNotFoundWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	err := wiRepo.DeleteWorkoutImage(ctx, uuid.New(), userID)
	if !errors.Is(err, ErrWorkoutImageNotFound) {
		t.Errorf("Expected ErrWorkoutImageNotFound, but got %v", err)
	}
}

func TestDeleteWorkoutImageOnDeleteWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	wiRepo := NewWorkoutImageRepository(db)
	workoutRepo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, _ := testutil.InsertProfile(ctx, db, "testuser")
	workout, _ := workoutRepo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	wi, _ := wiRepo.CreateWorkoutImage(ctx, workout.ID, "/path/to/delete.jpg", 0, userID)
	err := workoutRepo.DeleteWorkout(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete workout: %v", err)
	}

	_, err = wiRepo.GetWorkoutImageByID(ctx, wi.ID)
	if !errors.Is(err, ErrWorkoutImageNotFound) {
		t.Errorf("Expected ErrWorkoutImageNotFound, but got %v", err)
	}
}

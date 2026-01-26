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

func TestCreateWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Morning Workout"
	startedAt := time.Now()
	endedAt := time.Now().Add(time.Hour)

	workout, err := repo.Create(ctx, userID, &name, nil, startedAt, endedAt, 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	expectedTotalWeight := 360
	_, err = db.Exec(
		ctx,
		"UPDATE public.workouts SET total_weight = $1 WHERE id = $2",
		expectedTotalWeight,
		workout.ID,
	)
	if err != nil {
		t.Fatalf("Failed to update workout: %v", err)
	}

	if workout.ID == uuid.Nil {
		t.Error("Workout ID should not be nil")
	}
	if workout.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", workout.UserID, userID)
	}
	if *workout.Name != name {
		t.Errorf("Name mismatch: got %v, want %v", *workout.Name, name)
	}
	if workout.LikesCount != 0 {
		t.Errorf("LikesCount should be 0: got %v", workout.LikesCount)
	}
	if workout.CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}

	var workoutCount int
	var totalWeight int
	err = db.QueryRow(
		ctx,
		"SELECT total_workouts, total_weight FROM public.profiles WHERE id = $1",
		userID,
	).Scan(&workoutCount, &totalWeight)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	if workoutCount != 1 {
		t.Errorf("Expected 1 workout, got %d", workoutCount)
	}
	if totalWeight != expectedTotalWeight {
		t.Errorf("TotalWeight should be %d: got %v", expectedTotalWeight, totalWeight)
	}
}

func TestCreateWorkoutWithNullName(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	workout, err := repo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	if workout.Name != nil {
		t.Errorf("Name should be nil: got %v", workout.Name)
	}
}

func TestNotFoundUserCreateWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, uuid.New(), nil, nil, time.Now(), time.Now(), 0)
	if err == nil {
		t.Error("Expected error when user is not found, but got nil")
	}
	if !errors.Is(err, ErrReferenceViolation) {
		t.Errorf("Expected ErrReferenceViolation, but got %v", err)
	}
}

func TestGetWorkoutByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Test Workout"
	created, err := repo.Create(ctx, userID, &name, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	workout, err := repo.GetWorkoutByID(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get workout: %v", err)
	}

	if workout.ID != created.ID {
		t.Errorf("ID mismatch: got %v, want %v", workout.ID, created.ID)
	}
}

func TestGetWorkoutByIDNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = repo.GetWorkoutByID(ctx, uuid.New(), userID)
	if err == nil {
		t.Error("Expected error when workout is not found, but got nil")
	}

	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Errorf("Expected ErrWorkoutNotFound, but got %v", err)
	}
}

func TestGetWorkoutByIDBlockedUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
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
		"INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		userID,
		blockedUserID,
	)
	if err != nil {
		t.Fatalf("Failed to insert blocked user: %v", err)
	}

	_, err = repo.GetWorkoutByID(ctx, uuid.New(), blockedUserID)
	if err == nil {
		t.Error("Expected error when workout is not found, but got nil")
	}

	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Errorf("Expected ErrWorkoutNotFound, but got %v", err)
	}
}

func TestGetWorkoutsByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name1 := "Workout 1"
	name2 := "Workout 2"
	_, err = repo.Create(ctx, userID, &name1, nil, time.Now().Add(-time.Hour), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}
	_, err = repo.Create(ctx, userID, &name2, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	workouts, err := repo.GetWorkoutsByUserID(ctx, userID, userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get workouts: %v", err)
	}

	if len(workouts) != 2 {
		t.Errorf("Expected 2 workouts, got %d", len(workouts))
	}

	// Should be ordered by started_at DESC (most recent first)
	if *workouts[0].Name != name2 {
		t.Errorf("Expected first workout to be '%s', got %s", name2, *workouts[0].Name)
	}
}

func TestGetWorkoutsByUserIDWithPagination(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	for i := 0; i < 5; i++ {
		name := "Workout"
		_, err = repo.Create(ctx, userID, &name, nil, time.Now().Add(time.Duration(i)*time.Hour), time.Now(), 0)
		if err != nil {
			t.Fatalf("Failed to create workout: %v", err)
		}
	}

	workouts, err := repo.GetWorkoutsByUserID(ctx, userID, userID, 2, 0)
	if err != nil {
		t.Fatalf("Failed to get workouts: %v", err)
	}

	if len(workouts) != 2 {
		t.Errorf("Expected 2 workouts with limit, got %d", len(workouts))
	}

	workouts, err = repo.GetWorkoutsByUserID(ctx, userID, userID, 2, 2)
	if err != nil {
		t.Fatalf("Failed to get workouts: %v", err)
	}

	if len(workouts) != 2 {
		t.Errorf("Expected 2 workouts with offset, got %d", len(workouts))
	}
}

func TestGetWorkoutsByUserIDBlockedUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = repo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	blockedUserID, _, err := testutil.InsertProfile(ctx, db, "blockeduser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	_, err = db.Exec(
		ctx,
		"INSERT INTO public.blocked_users (blocker_id, blocked_id) VALUES ($1, $2)",
		userID,
		blockedUserID,
	)
	if err != nil {
		t.Fatalf("Failed to insert blocked user: %v", err)
	}

	workouts, err := repo.GetWorkoutsByUserID(ctx, userID, blockedUserID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get workouts: %v", err)
	}

	if len(workouts) != 0 {
		t.Errorf("Expected 0 workouts, got %d", len(workouts))
	}
}

func TestUpdateWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Original"
	workout, err := repo.Create(ctx, userID, &name, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	newName := "Updated Workout"
	comment := "Great workout!"
	endedAt := time.Now()
	err = repo.UpdateWorkout(ctx, workout.ID, models.UpdateWorkoutRequest{
		Name:    &newName,
		Comment: &comment,
		EndedAt: &endedAt,
	})
	if err != nil {
		t.Fatalf("Failed to update workout: %v", err)
	}

	updated, err := repo.GetWorkoutByID(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get workout: %v", err)
	}

	if *updated.Name != newName {
		t.Errorf("Name was not updated: got %v, want %v", *updated.Name, newName)
	}
	if *updated.Comment != comment {
		t.Errorf("Comment was not updated: got %v, want %v", *updated.Comment, comment)
	}
	if updated.EndedAt.IsZero() {
		t.Error("EndedAt should not be nil")
	}
}

func TestUpdateWorkoutNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	newName := "Test"
	err := repo.UpdateWorkout(ctx, uuid.New(), models.UpdateWorkoutRequest{
		Name: &newName,
	})
	if err == nil {
		t.Error("Expected error when workout is not found, but got nil")
	}

	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Errorf("Expected ErrWorkoutNotFound, but got %v", err)
	}
}

func TestUpdateWorkoutNoChange(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Original"
	workout, err := repo.Create(ctx, userID, &name, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	err = repo.UpdateWorkout(ctx, workout.ID, models.UpdateWorkoutRequest{})
	if err != nil {
		t.Fatalf("Expected no error for empty update, got: %v", err)
	}

	updated, err := repo.GetWorkoutByID(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get workout: %v", err)
	}

	if updated.UpdatedAt != workout.UpdatedAt {
		t.Errorf("UpdatedAt should not have changed")
	}
}

func TestUpdateWorkoutZeroToNull(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	name := "Test"
	workout, err := repo.Create(ctx, userID, &name, nil, time.Now(), time.Now(), 0)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	// First set values
	comment := "Comment"
	duration := 3600
	err = repo.UpdateWorkout(ctx, workout.ID, models.UpdateWorkoutRequest{
		Comment:         &comment,
		DurationSeconds: &duration,
	})
	if err != nil {
		t.Fatalf("Failed to update workout: %v", err)
	}

	// Then set to zero/empty to trigger NULL conversion
	emptyComment := ""
	zeroDuration := 0
	err = repo.UpdateWorkout(ctx, workout.ID, models.UpdateWorkoutRequest{
		Comment:         &emptyComment,
		DurationSeconds: &zeroDuration,
	})
	if err != nil {
		t.Fatalf("Failed to update workout: %v", err)
	}

	updated, err := repo.GetWorkoutByID(ctx, workout.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get workout: %v", err)
	}

	if updated.Comment != nil {
		t.Errorf("Comment should be nil: got %v", updated.Comment)
	}
	if updated.DurationSeconds != 0 {
		t.Errorf("DurationSeconds should be nil: got %v", updated.DurationSeconds)
	}
}

func TestDeleteWorkout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	userID, _, err := testutil.InsertProfile(ctx, db, "testuser")
	if err != nil {
		t.Fatalf("Failed to insert profile: %v", err)
	}

	profileTotalWeight := 360
	_, err = db.Exec(
		ctx,
		"UPDATE public.profiles SET total_weight = $1 WHERE id = $2",
		profileTotalWeight,
		userID,
	)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	workout, err := repo.Create(ctx, userID, nil, nil, time.Now(), time.Now(), 180)
	if err != nil {
		t.Fatalf("Failed to create workout: %v", err)
	}

	workoutID := workout.ID

	workoutTotalWeight := 180
	_, err = db.Exec(
		ctx,
		"UPDATE public.workouts SET total_weight = $1 WHERE id = $2",
		workoutTotalWeight,
		workoutID,
	)
	if err != nil {
		t.Fatalf("Failed to update workout: %v", err)
	}

	err = repo.DeleteWorkout(ctx, workoutID)
	if err != nil {
		t.Fatalf("Failed to delete workout: %v", err)
	}

	_, err = repo.GetWorkoutByID(ctx, workoutID, userID)
	if err == nil {
		t.Error("Expected workout to be deleted, but got nil")
	}
	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Errorf("Expected ErrWorkoutNotFound, but got %v", err)
	}

	var workoutCount int
	var totalWeight int
	err = db.QueryRow(
		ctx,
		"SELECT total_workouts, total_weight FROM public.profiles WHERE id = $1",
		userID,
	).Scan(&workoutCount, &totalWeight)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	if workoutCount != 0 {
		t.Errorf("Expected 0 workouts, got %d", workoutCount)
	}
	if totalWeight != profileTotalWeight {
		t.Errorf("TotalWeight should be %d: got %v", profileTotalWeight, totalWeight)
	}
}

func TestDeleteWorkoutNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()
	repo := NewWorkoutRepository(db)
	ctx := context.Background()

	err := repo.DeleteWorkout(ctx, uuid.New())
	if err == nil {
		t.Error("Expected error when workout is not found, but got nil")
	}

	if !errors.Is(err, ErrWorkoutNotFound) {
		t.Errorf("Expected ErrWorkoutNotFound, but got %v", err)
	}
}

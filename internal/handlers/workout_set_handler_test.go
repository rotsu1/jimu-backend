package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/handlers/testutils"
	"github.com/rotsu1/jimu-backend/internal/models"
)

// --- Mocks ---

type mockWorkoutSetRepo struct {
	CreateWorkoutSetFunc                  func(ctx context.Context, workoutExerciseID uuid.UUID, weight *float64, reps *int, isCompleted bool, orderIndex int, userID uuid.UUID) (*models.WorkoutSet, error)
	GetWorkoutSetByIDFunc                 func(ctx context.Context, id uuid.UUID) (*models.WorkoutSet, error)
	GetWorkoutSetsByWorkoutExerciseIDFunc func(ctx context.Context, workoutExerciseID uuid.UUID) ([]*models.WorkoutSet, error)
	UpdateWorkoutSetFunc                  func(ctx context.Context, workoutSetID uuid.UUID, userID uuid.UUID, updates models.UpdateWorkoutSetRequest) error
	DeleteWorkoutSetFunc                  func(ctx context.Context, workoutSetID uuid.UUID, userID uuid.UUID) error
}

func (m *mockWorkoutSetRepo) CreateWorkoutSet(ctx context.Context, workoutExerciseID uuid.UUID, weight *float64, reps *int, isCompleted bool, orderIndex int, userID uuid.UUID) (*models.WorkoutSet, error) {
	if m.CreateWorkoutSetFunc != nil {
		return m.CreateWorkoutSetFunc(ctx, workoutExerciseID, weight, reps, isCompleted, orderIndex, userID)
	}
	return &models.WorkoutSet{ID: uuid.New(), WorkoutExerciseID: workoutExerciseID}, nil
}

func (m *mockWorkoutSetRepo) GetWorkoutSetByID(ctx context.Context, id uuid.UUID) (*models.WorkoutSet, error) {
	if m.GetWorkoutSetByIDFunc != nil {
		return m.GetWorkoutSetByIDFunc(ctx, id)
	}
	return &models.WorkoutSet{ID: id}, nil
}

func (m *mockWorkoutSetRepo) GetWorkoutSetsByWorkoutExerciseID(ctx context.Context, workoutExerciseID uuid.UUID) ([]*models.WorkoutSet, error) {
	if m.GetWorkoutSetsByWorkoutExerciseIDFunc != nil {
		return m.GetWorkoutSetsByWorkoutExerciseIDFunc(ctx, workoutExerciseID)
	}
	return []*models.WorkoutSet{}, nil
}

func (m *mockWorkoutSetRepo) UpdateWorkoutSet(ctx context.Context, workoutSetID uuid.UUID, userID uuid.UUID, updates models.UpdateWorkoutSetRequest) error {
	if m.UpdateWorkoutSetFunc != nil {
		return m.UpdateWorkoutSetFunc(ctx, workoutSetID, userID, updates)
	}
	return nil
}

func (m *mockWorkoutSetRepo) DeleteWorkoutSet(ctx context.Context, workoutSetID uuid.UUID, userID uuid.UUID) error {
	if m.DeleteWorkoutSetFunc != nil {
		return m.DeleteWorkoutSetFunc(ctx, workoutSetID, userID)
	}
	return nil
}

// --- Tests ---

func TestAddSetToWorkoutExercise_Success(t *testing.T) {
	h := NewWorkoutSetHandler(&mockWorkoutSetRepo{})

	body := `{"reps": 12, "weight": 60}`
	req := httptest.NewRequest("POST", "/workout-exercises/00000000-0000-0000-0000-000000000001/sets", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.AddSet(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestUpdateSetInWorkout_Success(t *testing.T) {
	h := NewWorkoutSetHandler(&mockWorkoutSetRepo{})

	body := `{"reps": 15}`
	req := httptest.NewRequest("PUT", "/workout-sets/00000000-0000-0000-0000-000000000001", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpdateSet(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestRemoveSetFromWorkout_Success(t *testing.T) {
	h := NewWorkoutSetHandler(&mockWorkoutSetRepo{})

	req := httptest.NewRequest("DELETE", "/workout-sets/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveSet(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

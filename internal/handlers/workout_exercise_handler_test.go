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

type mockWorkoutExerciseRepo struct {
	CreateWorkoutExerciseFunc          func(ctx context.Context, workoutID uuid.UUID, exerciseID uuid.UUID, orderIndex int, memo *string, restTimerSeconds *int, userID uuid.UUID) (*models.WorkoutExercise, error)
	GetWorkoutExerciseByIDFunc         func(ctx context.Context, id uuid.UUID) (*models.WorkoutExercise, error)
	GetWorkoutExercisesByWorkoutIDFunc func(ctx context.Context, workoutID uuid.UUID) ([]*models.WorkoutExercise, error)
	UpdateWorkoutExerciseFunc          func(ctx context.Context, workoutExerciseID uuid.UUID, updates models.UpdateWorkoutExerciseRequest, userID uuid.UUID) error
	DeleteWorkoutExerciseFunc          func(ctx context.Context, workoutExerciseID uuid.UUID, userID uuid.UUID) error
}

func (m *mockWorkoutExerciseRepo) CreateWorkoutExercise(ctx context.Context, workoutID uuid.UUID, exerciseID uuid.UUID, orderIndex int, memo *string, restTimerSeconds *int, userID uuid.UUID) (*models.WorkoutExercise, error) {
	if m.CreateWorkoutExerciseFunc != nil {
		return m.CreateWorkoutExerciseFunc(ctx, workoutID, exerciseID, orderIndex, memo, restTimerSeconds, userID)
	}
	return &models.WorkoutExercise{ID: uuid.New(), WorkoutID: workoutID, ExerciseID: exerciseID}, nil
}

func (m *mockWorkoutExerciseRepo) GetWorkoutExerciseByID(ctx context.Context, id uuid.UUID) (*models.WorkoutExercise, error) {
	if m.GetWorkoutExerciseByIDFunc != nil {
		return m.GetWorkoutExerciseByIDFunc(ctx, id)
	}
	return &models.WorkoutExercise{ID: id}, nil
}

func (m *mockWorkoutExerciseRepo) GetWorkoutExercisesByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]*models.WorkoutExercise, error) {
	if m.GetWorkoutExercisesByWorkoutIDFunc != nil {
		return m.GetWorkoutExercisesByWorkoutIDFunc(ctx, workoutID)
	}
	return []*models.WorkoutExercise{}, nil
}

func (m *mockWorkoutExerciseRepo) UpdateWorkoutExercise(ctx context.Context, workoutExerciseID uuid.UUID, updates models.UpdateWorkoutExerciseRequest, userID uuid.UUID) error {
	if m.UpdateWorkoutExerciseFunc != nil {
		return m.UpdateWorkoutExerciseFunc(ctx, workoutExerciseID, updates, userID)
	}
	return nil
}

func (m *mockWorkoutExerciseRepo) DeleteWorkoutExercise(ctx context.Context, workoutExerciseID uuid.UUID, userID uuid.UUID) error {
	if m.DeleteWorkoutExerciseFunc != nil {
		return m.DeleteWorkoutExerciseFunc(ctx, workoutExerciseID, userID)
	}
	return nil
}

// --- Tests ---

func TestAddExerciseToWorkout_Success(t *testing.T) {
	h := NewWorkoutExerciseHandler(&mockWorkoutExerciseRepo{})

	body := `{"exercise_id": "00000000-0000-0000-0000-000000000002"}`
	req := httptest.NewRequest("POST", "/workouts/00000000-0000-0000-0000-000000000001/exercises", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.AddExercise(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestRemoveExerciseFromWorkout_Success(t *testing.T) {
	h := NewWorkoutExerciseHandler(&mockWorkoutExerciseRepo{})

	req := httptest.NewRequest("DELETE", "/workouts/00000000-0000-0000-0000-000000000001/exercises/00000000-0000-0000-0000-000000000003", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveExercise(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestUpdateExerciseInWorkout_Success(t *testing.T) {
	h := NewWorkoutExerciseHandler(&mockWorkoutExerciseRepo{})

	body := `{"memo": "Feeling strong"}`
	req := httptest.NewRequest("PUT", "/workouts/00000000-0000-0000-0000-000000000001/exercises/00000000-0000-0000-0000-000000000003", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpdateExercise(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/handlers/testutils"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

// --- Mocks ---

type mockRoutineExerciseRepo struct {
	CreateRoutineExerciseFunc          func(ctx context.Context, routineID uuid.UUID, exerciseID uuid.UUID, orderIndex *int, restTimerSeconds *int, memo *string, userID uuid.UUID) (*models.RoutineExercise, error)
	GetRoutineExerciseByIDFunc         func(ctx context.Context, id uuid.UUID) (*models.RoutineExercise, error)
	GetRoutineExercisesByRoutineIDFunc func(ctx context.Context, routineID uuid.UUID) ([]*models.RoutineExercise, error)
	UpdateRoutineExerciseFunc          func(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineExerciseRequest, userID uuid.UUID) error
	DeleteRoutineExerciseFunc          func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockRoutineExerciseRepo) CreateRoutineExercise(ctx context.Context, routineID uuid.UUID, exerciseID uuid.UUID, orderIndex *int, restTimerSeconds *int, memo *string, userID uuid.UUID) (*models.RoutineExercise, error) {
	if m.CreateRoutineExerciseFunc != nil {
		return m.CreateRoutineExerciseFunc(ctx, routineID, exerciseID, orderIndex, restTimerSeconds, memo, userID)
	}
	return &models.RoutineExercise{ID: uuid.New(), RoutineID: routineID, ExerciseID: exerciseID}, nil
}

func (m *mockRoutineExerciseRepo) GetRoutineExerciseByID(ctx context.Context, id uuid.UUID) (*models.RoutineExercise, error) {
	if m.GetRoutineExerciseByIDFunc != nil {
		return m.GetRoutineExerciseByIDFunc(ctx, id)
	}
	return &models.RoutineExercise{ID: id}, nil
}

func (m *mockRoutineExerciseRepo) GetRoutineExercisesByRoutineID(ctx context.Context, routineID uuid.UUID) ([]*models.RoutineExercise, error) {
	if m.GetRoutineExercisesByRoutineIDFunc != nil {
		return m.GetRoutineExercisesByRoutineIDFunc(ctx, routineID)
	}
	return []*models.RoutineExercise{}, nil
}

func (m *mockRoutineExerciseRepo) UpdateRoutineExercise(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineExerciseRequest, userID uuid.UUID) error {
	if m.UpdateRoutineExerciseFunc != nil {
		return m.UpdateRoutineExerciseFunc(ctx, id, updates, userID)
	}
	return nil
}

func (m *mockRoutineExerciseRepo) DeleteRoutineExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if m.DeleteRoutineExerciseFunc != nil {
		return m.DeleteRoutineExerciseFunc(ctx, id, userID)
	}
	return nil
}

// --- Tests ---

func TestAddExerciseToRoutine_Success(t *testing.T) {
	h := NewRoutineExerciseHandler(&mockRoutineExerciseRepo{})

	body := `{"exercise_id": "00000000-0000-0000-0000-000000000002"}`

	// ADD THE QUERY PARAM: ?id=...
	// This ensures GetIDFromRequest(r) finds the Routine ID
	targetID := "00000000-0000-0000-0000-000000000001"
	url := fmt.Sprintf("/routines/exercises?id=%s", targetID)

	req := httptest.NewRequest("POST", url, strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.AddExercise(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestRemoveExerciseFromRoutine_Success(t *testing.T) {
	h := NewRoutineExerciseHandler(&mockRoutineExerciseRepo{})

	req := httptest.NewRequest("DELETE", "/routines/00000000-0000-0000-0000-000000000001/exercises/00000000-0000-0000-0000-000000000003", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveExercise(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestRemoveExerciseFromRoutine_NotFound(t *testing.T) {
	mockRepo := &mockRoutineExerciseRepo{
		DeleteRoutineExerciseFunc: func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
			return repository.ErrRoutineExerciseNotFound
		},
	}
	h := NewRoutineExerciseHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/routines/00000000-0000-0000-0000-000000000001/exercises/00000000-0000-0000-0000-000000000003", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveExercise(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

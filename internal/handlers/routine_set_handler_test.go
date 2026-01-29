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
	"github.com/rotsu1/jimu-backend/internal/repository"
)

// --- Mocks ---

type mockRoutineSetRepo struct {
	CreateRoutineSetFunc                  func(ctx context.Context, routineExerciseID uuid.UUID, weight *float64, reps *int, orderIndex int, userID uuid.UUID) (*models.RoutineSet, error)
	GetRoutineSetByIDFunc                 func(ctx context.Context, id uuid.UUID) (*models.RoutineSet, error)
	GetRoutineSetsByRoutineExerciseIDFunc func(ctx context.Context, routineExerciseID uuid.UUID) ([]*models.RoutineSet, error)
	UpdateRoutineSetFunc                  func(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineSetRequest, userID uuid.UUID) error
	DeleteRoutineSetFunc                  func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockRoutineSetRepo) CreateRoutineSet(ctx context.Context, routineExerciseID uuid.UUID, weight *float64, reps *int, orderIndex int, userID uuid.UUID) (*models.RoutineSet, error) {
	if m.CreateRoutineSetFunc != nil {
		return m.CreateRoutineSetFunc(ctx, routineExerciseID, weight, reps, orderIndex, userID)
	}
	return &models.RoutineSet{ID: uuid.New(), RoutineExerciseID: routineExerciseID}, nil
}

func (m *mockRoutineSetRepo) GetRoutineSetByID(ctx context.Context, id uuid.UUID) (*models.RoutineSet, error) {
	if m.GetRoutineSetByIDFunc != nil {
		return m.GetRoutineSetByIDFunc(ctx, id)
	}
	return &models.RoutineSet{ID: id}, nil
}

func (m *mockRoutineSetRepo) GetRoutineSetsByRoutineExerciseID(ctx context.Context, routineExerciseID uuid.UUID) ([]*models.RoutineSet, error) {
	if m.GetRoutineSetsByRoutineExerciseIDFunc != nil {
		return m.GetRoutineSetsByRoutineExerciseIDFunc(ctx, routineExerciseID)
	}
	return []*models.RoutineSet{}, nil
}

func (m *mockRoutineSetRepo) UpdateRoutineSet(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineSetRequest, userID uuid.UUID) error {
	if m.UpdateRoutineSetFunc != nil {
		return m.UpdateRoutineSetFunc(ctx, id, updates, userID)
	}
	return nil
}

func (m *mockRoutineSetRepo) DeleteRoutineSet(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if m.DeleteRoutineSetFunc != nil {
		return m.DeleteRoutineSetFunc(ctx, id, userID)
	}
	return nil
}

// --- Tests ---

func TestAddSetToRoutineExercise_Success(t *testing.T) {
	h := NewRoutineSetHandler(&mockRoutineSetRepo{})

	body := `{"reps": 10}`
	// Using path typical for nested structure or just mapping
	req := httptest.NewRequest("POST", "/routine-exercises/00000000-0000-0000-0000-000000000001/sets", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.AddSet(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestRemoveSetFromRoutineExercise_Success(t *testing.T) {
	h := NewRoutineSetHandler(&mockRoutineSetRepo{})

	req := httptest.NewRequest("DELETE", "/routine-sets/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveSet(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestRemoveSetFromRoutineExercise_NotFound(t *testing.T) {
	mockRepo := &mockRoutineSetRepo{
		DeleteRoutineSetFunc: func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
			return repository.ErrRoutineSetNotFound
		},
	}
	h := NewRoutineSetHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/routine-sets/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveSet(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

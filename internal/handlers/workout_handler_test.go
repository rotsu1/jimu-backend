package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/handlers/testutils"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

// --- Mocks ---

type mockWorkoutRepo struct {
	CreateFunc              func(ctx context.Context, userID uuid.UUID, name *string, comment *string, startedAt time.Time, endedAt time.Time, durationSeconds int) (*models.Workout, error)
	GetWorkoutByIDFunc      func(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID) (*models.Workout, error)
	GetWorkoutsByUserIDFunc func(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.Workout, error)
	UpdateWorkoutFunc       func(ctx context.Context, id uuid.UUID, updates models.UpdateWorkoutRequest, userID uuid.UUID) error
	DeleteWorkoutFunc       func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockWorkoutRepo) Create(ctx context.Context, userID uuid.UUID, name *string, comment *string, startedAt time.Time, endedAt time.Time, durationSeconds int) (*models.Workout, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, name, comment, startedAt, endedAt, durationSeconds)
	}
	return &models.Workout{ID: uuid.New(), UserID: userID}, nil
}

func (m *mockWorkoutRepo) GetWorkoutByID(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID) (*models.Workout, error) {
	if m.GetWorkoutByIDFunc != nil {
		return m.GetWorkoutByIDFunc(ctx, workoutID, viewerID)
	}
	return &models.Workout{ID: workoutID}, nil
}

func (m *mockWorkoutRepo) GetWorkoutsByUserID(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.Workout, error) {
	if m.GetWorkoutsByUserIDFunc != nil {
		return m.GetWorkoutsByUserIDFunc(ctx, targetID, viewerID, limit, offset)
	}
	return []*models.Workout{}, nil
}

func (m *mockWorkoutRepo) UpdateWorkout(ctx context.Context, id uuid.UUID, updates models.UpdateWorkoutRequest, userID uuid.UUID) error {
	if m.UpdateWorkoutFunc != nil {
		return m.UpdateWorkoutFunc(ctx, id, updates, userID)
	}
	return nil
}

func (m *mockWorkoutRepo) DeleteWorkout(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if m.DeleteWorkoutFunc != nil {
		return m.DeleteWorkoutFunc(ctx, id, userID)
	}
	return nil
}

// --- Tests ---

func TestCreateWorkout_Success(t *testing.T) {
	h := NewWorkoutHandler(&mockWorkoutRepo{})

	body := `{"name": "Morning Workout", "duration_seconds": 3600}`
	req := httptest.NewRequest("POST", "/workouts", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.CreateWorkout(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestGetWorkout_Success(t *testing.T) {
	h := NewWorkoutHandler(&mockWorkoutRepo{})

	req := httptest.NewRequest("GET", "/workouts/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetWorkout(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestListWorkouts_Success(t *testing.T) {
	h := NewWorkoutHandler(&mockWorkoutRepo{})

	req := httptest.NewRequest("GET", "/workouts", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListWorkouts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestUpdateWorkout_Success(t *testing.T) {
	h := NewWorkoutHandler(&mockWorkoutRepo{})

	body := `{"name": "Updated Name"}`
	req := httptest.NewRequest("PUT", "/workouts/00000000-0000-0000-0000-000000000001", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpdateWorkout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteWorkout_Success(t *testing.T) {
	h := NewWorkoutHandler(&mockWorkoutRepo{})

	req := httptest.NewRequest("DELETE", "/workouts/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteWorkout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteWorkout_NotFound(t *testing.T) {
	mockRepo := &mockWorkoutRepo{
		DeleteWorkoutFunc: func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
			return repository.ErrWorkoutNotFound
		},
	}
	h := NewWorkoutHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/workouts/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteWorkout(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

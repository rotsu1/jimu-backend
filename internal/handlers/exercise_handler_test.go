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

type mockExerciseRepo struct {
	CreateExerciseFunc       func(ctx context.Context, userID *uuid.UUID, name string, suggestedRestSeconds *int, icon *string, requesterID uuid.UUID) (*models.Exercise, error)
	GetExerciseByIDFunc      func(ctx context.Context, exerciseID uuid.UUID, userID uuid.UUID) (*models.Exercise, error)
	GetExercisesByUserIDFunc func(ctx context.Context, viewerID uuid.UUID, targetID uuid.UUID) ([]*models.Exercise, error)
	UpdateExerciseFunc       func(ctx context.Context, id uuid.UUID, updates models.UpdateExerciseRequest, userID uuid.UUID) error
	DeleteExerciseFunc       func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockExerciseRepo) CreateExercise(ctx context.Context, userID *uuid.UUID, name string, suggestedRestSeconds *int, icon *string, requesterID uuid.UUID) (*models.Exercise, error) {
	if m.CreateExerciseFunc != nil {
		return m.CreateExerciseFunc(ctx, userID, name, suggestedRestSeconds, icon, requesterID)
	}
	return &models.Exercise{ID: uuid.New()}, nil
}

func (m *mockExerciseRepo) GetExerciseByID(ctx context.Context, exerciseID uuid.UUID, userID uuid.UUID) (*models.Exercise, error) {
	if m.GetExerciseByIDFunc != nil {
		return m.GetExerciseByIDFunc(ctx, exerciseID, userID)
	}
	return &models.Exercise{ID: exerciseID}, nil
}

func (m *mockExerciseRepo) GetExercisesByUserID(ctx context.Context, viewerID uuid.UUID, targetID uuid.UUID) ([]*models.Exercise, error) {
	if m.GetExercisesByUserIDFunc != nil {
		return m.GetExercisesByUserIDFunc(ctx, viewerID, targetID)
	}
	return []*models.Exercise{}, nil
}

func (m *mockExerciseRepo) UpdateExercise(ctx context.Context, id uuid.UUID, updates models.UpdateExerciseRequest, userID uuid.UUID) error {
	if m.UpdateExerciseFunc != nil {
		return m.UpdateExerciseFunc(ctx, id, updates, userID)
	}
	return nil
}

func (m *mockExerciseRepo) DeleteExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if m.DeleteExerciseFunc != nil {
		return m.DeleteExerciseFunc(ctx, id, userID)
	}
	return nil
}

// --- Tests ---

func TestCreateExercise_Success(t *testing.T) {
	h := NewExerciseHandler(&mockExerciseRepo{})

	body := `{"name": "Bench Press", "suggested_rest_seconds": 120, "icon": "bench"}`
	req := httptest.NewRequest("POST", "/exercises", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.CreateExercise(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestCreateExercise_InvalidInput(t *testing.T) {
	h := NewExerciseHandler(&mockExerciseRepo{})

	body := `invalid-json`
	req := httptest.NewRequest("POST", "/exercises", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.CreateExercise(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestGetExercise_Success(t *testing.T) {
	h := NewExerciseHandler(&mockExerciseRepo{})

	req := httptest.NewRequest("GET", "/exercises?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetExercise(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestGetExercise_NotFound(t *testing.T) {
	mockRepo := &mockExerciseRepo{
		GetExerciseByIDFunc: func(ctx context.Context, exerciseID uuid.UUID, userID uuid.UUID) (*models.Exercise, error) {
			return nil, repository.ErrExerciseNotFound
		},
	}
	h := NewExerciseHandler(mockRepo)

	req := httptest.NewRequest("GET", "/exercises?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetExercise(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

func TestListExercises_Success(t *testing.T) {
	h := NewExerciseHandler(&mockExerciseRepo{})

	req := httptest.NewRequest("GET", "/exercises", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListExercises(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestUpdateExercise_Success(t *testing.T) {
	h := NewExerciseHandler(&mockExerciseRepo{})

	body := `{"name": "New Name"}`
	req := httptest.NewRequest("PUT", "/exercises?id=00000000-0000-0000-0000-000000000001", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpdateExercise(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteExercise_Success(t *testing.T) {
	h := NewExerciseHandler(&mockExerciseRepo{})

	req := httptest.NewRequest("DELETE", "/exercises?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteExercise(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

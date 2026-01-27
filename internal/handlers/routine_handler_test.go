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

type mockRoutineRepo struct {
	CreateRoutineFunc       func(ctx context.Context, userID uuid.UUID, name string) (*models.Routine, error)
	GetRoutineByIDFunc      func(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) (*models.Routine, error)
	GetRoutinesByUserIDFunc func(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID) ([]*models.Routine, error)
	UpdateRoutineFunc       func(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineRequest, userID uuid.UUID) error
	DeleteRoutineFunc       func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockRoutineRepo) CreateRoutine(ctx context.Context, userID uuid.UUID, name string) (*models.Routine, error) {
	if m.CreateRoutineFunc != nil {
		return m.CreateRoutineFunc(ctx, userID, name)
	}
	return &models.Routine{ID: uuid.New(), UserID: userID, Name: name}, nil
}

func (m *mockRoutineRepo) GetRoutineByID(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) (*models.Routine, error) {
	if m.GetRoutineByIDFunc != nil {
		return m.GetRoutineByIDFunc(ctx, id, viewerID)
	}
	return &models.Routine{ID: id}, nil
}

func (m *mockRoutineRepo) GetRoutinesByUserID(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID) ([]*models.Routine, error) {
	if m.GetRoutinesByUserIDFunc != nil {
		return m.GetRoutinesByUserIDFunc(ctx, userID, viewerID)
	}
	return []*models.Routine{}, nil
}

func (m *mockRoutineRepo) UpdateRoutine(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineRequest, userID uuid.UUID) error {
	if m.UpdateRoutineFunc != nil {
		return m.UpdateRoutineFunc(ctx, id, updates, userID)
	}
	return nil
}

func (m *mockRoutineRepo) DeleteRoutine(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if m.DeleteRoutineFunc != nil {
		return m.DeleteRoutineFunc(ctx, id, userID)
	}
	return nil
}

// --- Tests ---

func TestCreateRoutine_Success(t *testing.T) {
	h := NewRoutineHandler(&mockRoutineRepo{})

	body := `{"name": "Morning Routine"}`
	req := httptest.NewRequest("POST", "/routines", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.CreateRoutine(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestGetRoutine_Success(t *testing.T) {
	h := NewRoutineHandler(&mockRoutineRepo{})

	req := httptest.NewRequest("GET", "/routines?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetRoutine(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestListRoutines_Success(t *testing.T) {
	h := NewRoutineHandler(&mockRoutineRepo{})

	req := httptest.NewRequest("GET", "/routines", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListRoutines(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestUpdateRoutine_Success(t *testing.T) {
	h := NewRoutineHandler(&mockRoutineRepo{})

	body := `{"name": "New Name"}`
	req := httptest.NewRequest("PUT", "/routines?id=00000000-0000-0000-0000-000000000001", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpdateRoutine(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteRoutine_Success(t *testing.T) {
	h := NewRoutineHandler(&mockRoutineRepo{})

	req := httptest.NewRequest("DELETE", "/routines?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteRoutine(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteRoutine_NotFound(t *testing.T) {
	mockRepo := &mockRoutineRepo{
		DeleteRoutineFunc: func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
			return repository.ErrRoutineNotFound
		},
	}
	h := NewRoutineHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/routines?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteRoutine(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

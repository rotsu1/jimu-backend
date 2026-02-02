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

type mockMuscleRepo struct {
	GetAllMusclesFunc   func(ctx context.Context) ([]*models.Muscle, error)
	GetMuscleByIDFunc   func(ctx context.Context, id uuid.UUID) (*models.Muscle, error)
	GetMuscleByNameFunc func(ctx context.Context, name string) (*models.Muscle, error)
	CreateMuscleFunc    func(ctx context.Context, name string, userID uuid.UUID) (*models.Muscle, error)
	DeleteMuscleFunc    func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockMuscleRepo) GetAllMuscles(ctx context.Context) ([]*models.Muscle, error) {
	if m.GetAllMusclesFunc != nil {
		return m.GetAllMusclesFunc(ctx)
	}
	return []*models.Muscle{}, nil
}

func (m *mockMuscleRepo) GetMuscleByID(ctx context.Context, id uuid.UUID) (*models.Muscle, error) {
	if m.GetMuscleByIDFunc != nil {
		return m.GetMuscleByIDFunc(ctx, id)
	}
	return &models.Muscle{ID: id}, nil
}

func (m *mockMuscleRepo) GetMuscleByName(ctx context.Context, name string) (*models.Muscle, error) {
	if m.GetMuscleByNameFunc != nil {
		return m.GetMuscleByNameFunc(ctx, name)
	}
	return &models.Muscle{Name: name}, nil
}

func (m *mockMuscleRepo) CreateMuscle(ctx context.Context, name string, userID uuid.UUID) (*models.Muscle, error) {
	if m.CreateMuscleFunc != nil {
		return m.CreateMuscleFunc(ctx, name, userID)
	}
	return &models.Muscle{Name: name, ID: uuid.New()}, nil
}

func (m *mockMuscleRepo) DeleteMuscle(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if m.DeleteMuscleFunc != nil {
		return m.DeleteMuscleFunc(ctx, id, userID)
	}
	return nil
}

// --- Tests ---

func TestListMuscles_Success(t *testing.T) {
	h := NewMuscleHandler(&mockMuscleRepo{})

	req := httptest.NewRequest("GET", "/muscles", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListMuscles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestGetMuscle_Success(t *testing.T) {
	h := NewMuscleHandler(&mockMuscleRepo{})

	req := httptest.NewRequest("GET", "/muscles/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetMuscle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestCreateMuscle_Success(t *testing.T) {
	h := NewMuscleHandler(&mockMuscleRepo{})

	body := `{"name": "Chest"}`
	req := httptest.NewRequest("POST", "/muscles", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.CreateMuscle(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestDeleteMuscle_Success(t *testing.T) {
	h := NewMuscleHandler(&mockMuscleRepo{})

	req := httptest.NewRequest("DELETE", "/muscles/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteMuscle(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteMuscle_Unauthorized(t *testing.T) {
	mockRepo := &mockMuscleRepo{
		DeleteMuscleFunc: func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
			return repository.ErrUnauthorizedAction
		},
	}
	h := NewMuscleHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/muscles/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteMuscle(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", rr.Code)
	}
}

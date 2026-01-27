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

type mockExerciseTargetMuscleRepo struct {
	GetByExerciseIDFunc    func(ctx context.Context, exerciseID uuid.UUID) ([]*models.ExerciseTargetMuscle, error)
	AddTargetMuscleFunc    func(ctx context.Context, exerciseID uuid.UUID, muscleID uuid.UUID, userID uuid.UUID) (*models.ExerciseTargetMuscle, error)
	RemoveTargetMuscleFunc func(ctx context.Context, exerciseID uuid.UUID, muscleID uuid.UUID, userID uuid.UUID) error
}

func (m *mockExerciseTargetMuscleRepo) GetByExerciseID(ctx context.Context, exerciseID uuid.UUID) ([]*models.ExerciseTargetMuscle, error) {
	if m.GetByExerciseIDFunc != nil {
		return m.GetByExerciseIDFunc(ctx, exerciseID)
	}
	return []*models.ExerciseTargetMuscle{}, nil
}

func (m *mockExerciseTargetMuscleRepo) AddTargetMuscle(ctx context.Context, exerciseID uuid.UUID, muscleID uuid.UUID, userID uuid.UUID) (*models.ExerciseTargetMuscle, error) {
	if m.AddTargetMuscleFunc != nil {
		return m.AddTargetMuscleFunc(ctx, exerciseID, muscleID, userID)
	}
	return &models.ExerciseTargetMuscle{ExerciseID: exerciseID, MuscleID: muscleID}, nil
}

func (m *mockExerciseTargetMuscleRepo) RemoveTargetMuscle(ctx context.Context, exerciseID uuid.UUID, muscleID uuid.UUID, userID uuid.UUID) error {
	if m.RemoveTargetMuscleFunc != nil {
		return m.RemoveTargetMuscleFunc(ctx, exerciseID, muscleID, userID)
	}
	return nil
}

// --- Tests ---

func TestAddTargetMuscle_Success(t *testing.T) {
	h := NewExerciseTargetMuscleHandler(&mockExerciseTargetMuscleRepo{})

	body := `{"muscle_id": "00000000-0000-0000-0000-000000000002"}`
	req := httptest.NewRequest("POST", "/exercises/00000000-0000-0000-0000-000000000001/muscles", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.AddTargetMuscle(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestRemoveTargetMuscle_Success(t *testing.T) {
	h := NewExerciseTargetMuscleHandler(&mockExerciseTargetMuscleRepo{})

	req := httptest.NewRequest("DELETE", "/exercises/00000000-0000-0000-0000-000000000001/muscles/00000000-0000-0000-0000-000000000002", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveTargetMuscle(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestRemoveTargetMuscle_NotFound(t *testing.T) {
	mockRepo := &mockExerciseTargetMuscleRepo{
		RemoveTargetMuscleFunc: func(ctx context.Context, exerciseID uuid.UUID, muscleID uuid.UUID, userID uuid.UUID) error {
			return repository.ErrNotFound
		},
	}
	h := NewExerciseTargetMuscleHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/exercises/00000000-0000-0000-0000-000000000001/muscles/00000000-0000-0000-0000-000000000002", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveTargetMuscle(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

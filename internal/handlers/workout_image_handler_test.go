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

type mockWorkoutImageRepo struct {
	CreateWorkoutImageFunc          func(ctx context.Context, workoutID uuid.UUID, storagePath string, displayOrder *int, userID uuid.UUID) (*models.WorkoutImage, error)
	GetWorkoutImageByIDFunc         func(ctx context.Context, id uuid.UUID) (*models.WorkoutImage, error)
	GetWorkoutImagesByWorkoutIDFunc func(ctx context.Context, workoutID uuid.UUID) ([]*models.WorkoutImage, error)
	DeleteWorkoutImageFunc          func(ctx context.Context, workoutImageID uuid.UUID, userID uuid.UUID) error
}

func (m *mockWorkoutImageRepo) CreateWorkoutImage(ctx context.Context, workoutID uuid.UUID, storagePath string, displayOrder *int, userID uuid.UUID) (*models.WorkoutImage, error) {
	if m.CreateWorkoutImageFunc != nil {
		return m.CreateWorkoutImageFunc(ctx, workoutID, storagePath, displayOrder, userID)
	}
	return &models.WorkoutImage{ID: uuid.New(), WorkoutID: workoutID, StoragePath: storagePath}, nil
}

func (m *mockWorkoutImageRepo) GetWorkoutImageByID(ctx context.Context, id uuid.UUID) (*models.WorkoutImage, error) {
	if m.GetWorkoutImageByIDFunc != nil {
		return m.GetWorkoutImageByIDFunc(ctx, id)
	}
	return &models.WorkoutImage{ID: id}, nil
}

func (m *mockWorkoutImageRepo) GetWorkoutImagesByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]*models.WorkoutImage, error) {
	if m.GetWorkoutImagesByWorkoutIDFunc != nil {
		return m.GetWorkoutImagesByWorkoutIDFunc(ctx, workoutID)
	}
	return []*models.WorkoutImage{}, nil
}

func (m *mockWorkoutImageRepo) DeleteWorkoutImage(ctx context.Context, workoutImageID uuid.UUID, userID uuid.UUID) error {
	if m.DeleteWorkoutImageFunc != nil {
		return m.DeleteWorkoutImageFunc(ctx, workoutImageID, userID)
	}
	return nil
}

// --- Tests ---

func TestAddImageToWorkout_Success(t *testing.T) {
	h := NewWorkoutImageHandler(&mockWorkoutImageRepo{})

	body := `{"storage_path": "/path/to/image.jpg"}`
	req := httptest.NewRequest("POST", "/workouts/00000000-0000-0000-0000-000000000001/images", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.AddImage(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestRemoveImageFromWorkout_Success(t *testing.T) {
	h := NewWorkoutImageHandler(&mockWorkoutImageRepo{})

	req := httptest.NewRequest("DELETE", "/workout-images/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RemoveImage(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestListImagesForWorkout_Success(t *testing.T) {
	h := NewWorkoutImageHandler(&mockWorkoutImageRepo{})

	req := httptest.NewRequest("GET", "/workouts/00000000-0000-0000-0000-000000000001/images", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListImages(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

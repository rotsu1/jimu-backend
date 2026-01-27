package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/handlers/testutils"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

// --- Mocks ---

type mockWorkoutLikeRepo struct {
	LikeWorkoutFunc                func(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (*models.WorkoutLike, error)
	UnlikeWorkoutFunc              func(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) error
	GetWorkoutLikeByIDFunc         func(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (*models.WorkoutLike, error)
	GetWorkoutLikesByWorkoutIDFunc func(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.WorkoutLikeDetail, error)
	IsWorkoutLikedFunc             func(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (bool, error)
}

func (m *mockWorkoutLikeRepo) LikeWorkout(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (*models.WorkoutLike, error) {
	if m.LikeWorkoutFunc != nil {
		return m.LikeWorkoutFunc(ctx, userID, workoutID)
	}
	return &models.WorkoutLike{UserID: userID, WorkoutID: workoutID}, nil
}

func (m *mockWorkoutLikeRepo) UnlikeWorkout(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) error {
	if m.UnlikeWorkoutFunc != nil {
		return m.UnlikeWorkoutFunc(ctx, userID, workoutID)
	}
	return nil
}

func (m *mockWorkoutLikeRepo) GetWorkoutLikeByID(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (*models.WorkoutLike, error) {
	if m.GetWorkoutLikeByIDFunc != nil {
		return m.GetWorkoutLikeByIDFunc(ctx, userID, workoutID)
	}
	return &models.WorkoutLike{UserID: userID, WorkoutID: workoutID}, nil
}

func (m *mockWorkoutLikeRepo) GetWorkoutLikesByWorkoutID(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.WorkoutLikeDetail, error) {
	if m.GetWorkoutLikesByWorkoutIDFunc != nil {
		return m.GetWorkoutLikesByWorkoutIDFunc(ctx, workoutID, viewerID, limit, offset)
	}
	return []*models.WorkoutLikeDetail{}, nil
}

func (m *mockWorkoutLikeRepo) IsWorkoutLiked(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (bool, error) {
	if m.IsWorkoutLikedFunc != nil {
		return m.IsWorkoutLikedFunc(ctx, userID, workoutID)
	}
	return false, nil
}

// --- Tests ---

func TestLikeWorkout_Success(t *testing.T) {
	h := NewWorkoutLikeHandler(&mockWorkoutLikeRepo{})

	req := httptest.NewRequest("POST", "/workouts/00000000-0000-0000-0000-000000000001/likes", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.LikeWorkout(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestUnlikeWorkout_Success(t *testing.T) {
	h := NewWorkoutLikeHandler(&mockWorkoutLikeRepo{})

	req := httptest.NewRequest("DELETE", "/workouts/00000000-0000-0000-0000-000000000001/likes", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnlikeWorkout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestUnlikeWorkout_NotFound(t *testing.T) {
	mockRepo := &mockWorkoutLikeRepo{
		UnlikeWorkoutFunc: func(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) error {
			return repository.ErrWorkoutLikeNotFound
		},
	}
	h := NewWorkoutLikeHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/workouts/00000000-0000-0000-0000-000000000001/likes", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnlikeWorkout(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

func TestListWorkoutLikes_Success(t *testing.T) {
	h := NewWorkoutLikeHandler(&mockWorkoutLikeRepo{})

	req := httptest.NewRequest("GET", "/workouts/00000000-0000-0000-0000-000000000001/likes", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListLikes(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

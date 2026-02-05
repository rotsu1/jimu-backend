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
	CreateFunc               func(ctx context.Context, userID uuid.UUID, name *string, comment *string, startedAt time.Time, endedAt time.Time, durationSeconds int) (*models.Workout, error)
	GetWorkoutByIDFunc       func(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID) (*models.Workout, error)
	GetWorkoutsByUserIDFunc  func(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.Workout, error)
	GetTimelineWorkoutsFunc  func(ctx context.Context, viewerID uuid.UUID, targetID uuid.UUID, limit int, offset int) ([]*models.TimelineWorkout, error)
	UpdateWorkoutFunc        func(ctx context.Context, id uuid.UUID, updates models.UpdateWorkoutRequest, userID uuid.UUID) error
	DeleteWorkoutFunc        func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
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

func (m *mockWorkoutRepo) GetTimelineWorkouts(ctx context.Context, viewerID uuid.UUID, targetID uuid.UUID, limit int, offset int) ([]*models.TimelineWorkout, error) {
	if m.GetTimelineWorkoutsFunc != nil {
		return m.GetTimelineWorkoutsFunc(ctx, viewerID, targetID, limit, offset)
	}
	return []*models.TimelineWorkout{}, nil
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

func TestGetTimelineWorkouts(t *testing.T) {
	viewerID := uuid.New()

	tests := []struct {
		name           string
		url            string
		injectUserID   bool
		mockErr        error
		mockWorkouts   []*models.TimelineWorkout
		expectedStatus int
	}{
		{
			name:           "Success - returns 200 and timeline",
			url:            "/workouts/timeline",
			injectUserID:   true,
			mockWorkouts:   []*models.TimelineWorkout{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized - no user in context",
			url:            "/workouts/timeline",
			injectUserID:   false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid user_id query - 400",
			url:            "/workouts/timeline?user_id=not-a-uuid",
			injectUserID:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid limit - 400",
			url:            "/workouts/timeline?limit=abc",
			injectUserID:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid offset - 400",
			url:            "/workouts/timeline?offset=xyz",
			injectUserID:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "DB error - 500",
			url:            "/workouts/timeline",
			injectUserID:   true,
			mockErr:        context.DeadlineExceeded,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockWorkoutRepo{
				GetTimelineWorkoutsFunc: func(ctx context.Context, vID, tID uuid.UUID, limit, offset int) ([]*models.TimelineWorkout, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockWorkouts, nil
				},
			}
			h := NewWorkoutHandler(mockRepo)

			req := httptest.NewRequest("GET", tt.url, nil)
			if tt.injectUserID {
				req = testutils.InjectUserID(req, viewerID.String())
			}
			rr := httptest.NewRecorder()

			h.GetTimelineWorkouts(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestGetTimelineWorkouts_Success_WithTargetUser(t *testing.T) {
	viewerID := uuid.New()
	targetID := uuid.New()

	mockRepo := &mockWorkoutRepo{
		GetTimelineWorkoutsFunc: func(ctx context.Context, vID, tID uuid.UUID, limit, offset int) ([]*models.TimelineWorkout, error) {
			if tID != targetID || vID != viewerID {
				t.Errorf("repo called with targetID=%v viewerID=%v", tID, vID)
			}
			if limit != 10 || offset != 5 {
				t.Errorf("repo called with limit=%d offset=%d", limit, offset)
			}
			return []*models.TimelineWorkout{}, nil
		},
	}
	h := NewWorkoutHandler(mockRepo)

	req := httptest.NewRequest("GET", "/workouts/timeline?user_id="+targetID.String()+"&limit=10&offset=5", nil)
	req = testutils.InjectUserID(req, viewerID.String())
	rr := httptest.NewRecorder()

	h.GetTimelineWorkouts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

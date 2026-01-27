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

type mockCommentRepo struct {
	CreateCommentFunc          func(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID, parentID *uuid.UUID, content string) (*models.Comment, error)
	GetCommentByUserIDFunc     func(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) (*models.Comment, error)
	GetCommentsByWorkoutIDFunc func(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.Comment, error)
	GetRepliesFunc             func(ctx context.Context, commentID uuid.UUID, viewerID uuid.UUID) ([]*models.Comment, error)
	DeleteCommentFunc          func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockCommentRepo) CreateComment(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID, parentID *uuid.UUID, content string) (*models.Comment, error) {
	if m.CreateCommentFunc != nil {
		return m.CreateCommentFunc(ctx, userID, workoutID, parentID, content)
	}
	return &models.Comment{ID: uuid.New()}, nil
}

func (m *mockCommentRepo) GetCommentByUserID(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) (*models.Comment, error) {
	if m.GetCommentByUserIDFunc != nil {
		return m.GetCommentByUserIDFunc(ctx, id, viewerID)
	}
	return &models.Comment{ID: id}, nil
}

func (m *mockCommentRepo) GetCommentsByWorkoutID(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.Comment, error) {
	if m.GetCommentsByWorkoutIDFunc != nil {
		return m.GetCommentsByWorkoutIDFunc(ctx, workoutID, viewerID, limit, offset)
	}
	return []*models.Comment{}, nil
}

func (m *mockCommentRepo) GetReplies(ctx context.Context, commentID uuid.UUID, viewerID uuid.UUID) ([]*models.Comment, error) {
	if m.GetRepliesFunc != nil {
		return m.GetRepliesFunc(ctx, commentID, viewerID)
	}
	return []*models.Comment{}, nil
}

func (m *mockCommentRepo) DeleteComment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if m.DeleteCommentFunc != nil {
		return m.DeleteCommentFunc(ctx, id, userID)
	}
	return nil
}

// --- Tests ---

func TestCreateComment_Success(t *testing.T) {
	h := NewCommentHandler(&mockCommentRepo{})

	body := `{"workout_id": "00000000-0000-0000-0000-000000000001", "content": "Great job!"}`
	req := httptest.NewRequest("POST", "/comments", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.CreateComment(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestGetComment_Success(t *testing.T) {
	h := NewCommentHandler(&mockCommentRepo{})

	req := httptest.NewRequest("GET", "/comments?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetComment(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestListComments_ByWorkout(t *testing.T) {
	h := NewCommentHandler(&mockCommentRepo{})

	req := httptest.NewRequest("GET", "/comments?workout_id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListComments(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestDeleteComment_Success(t *testing.T) {
	h := NewCommentHandler(&mockCommentRepo{})

	req := httptest.NewRequest("DELETE", "/comments?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteComment(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteComment_MissingCommentID(t *testing.T) {
	h := NewCommentHandler(&mockCommentRepo{})

	req := httptest.NewRequest("DELETE", "/comments", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteComment(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

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

type mockCommentLikeRepo struct {
	LikeCommentFunc                func(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) (*models.CommentLike, error)
	UnlikeCommentFunc              func(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) error
	GetCommentLikesByCommentIDFunc func(ctx context.Context, commentID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.CommentLikeDetail, error)
}

func (m *mockCommentLikeRepo) LikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) (*models.CommentLike, error) {
	if m.LikeCommentFunc != nil {
		return m.LikeCommentFunc(ctx, userID, commentID)
	}
	return &models.CommentLike{UserID: userID, CommentID: commentID}, nil
}

func (m *mockCommentLikeRepo) UnlikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) error {
	if m.UnlikeCommentFunc != nil {
		return m.UnlikeCommentFunc(ctx, userID, commentID)
	}
	return nil
}

func (m *mockCommentLikeRepo) GetCommentLikesByCommentID(ctx context.Context, commentID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.CommentLikeDetail, error) {
	if m.GetCommentLikesByCommentIDFunc != nil {
		return m.GetCommentLikesByCommentIDFunc(ctx, commentID, viewerID, limit, offset)
	}
	return []*models.CommentLikeDetail{}, nil
}

// --- Tests ---

func TestLikeComment_Success(t *testing.T) {
	h := NewCommentLikeHandler(&mockCommentLikeRepo{})

	req := httptest.NewRequest("POST", "/comments/likes?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.LikeComment(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestUnlikeComment_Success(t *testing.T) {
	h := NewCommentLikeHandler(&mockCommentLikeRepo{})

	req := httptest.NewRequest("DELETE", "/comments/likes?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnlikeComment(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestUnlikeComment_NotFound(t *testing.T) {
	mockRepo := &mockCommentLikeRepo{
		UnlikeCommentFunc: func(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) error {
			return repository.ErrCommentLikeNotFound
		},
	}
	h := NewCommentLikeHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/comments/likes?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnlikeComment(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

func TestListCommentLikes_Success(t *testing.T) {
	h := NewCommentLikeHandler(&mockCommentLikeRepo{})

	req := httptest.NewRequest("GET", "/comments/likes?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListLikes(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

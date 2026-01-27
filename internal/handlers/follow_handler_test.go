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

type mockFollowRepo struct {
	FollowFunc          func(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*models.Follow, error)
	UnfollowFunc        func(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error
	GetFollowStatusFunc func(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*models.Follow, error)
	GetFollowersFunc    func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Follow, error)
	GetFollowingFunc    func(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Follow, error)
}

func (m *mockFollowRepo) Follow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*models.Follow, error) {
	if m.FollowFunc != nil {
		return m.FollowFunc(ctx, followerID, followingID)
	}
	return &models.Follow{FollowerID: followerID, FollowingID: followingID}, nil
}

func (m *mockFollowRepo) Unfollow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
	if m.UnfollowFunc != nil {
		return m.UnfollowFunc(ctx, followerID, followingID)
	}
	return nil
}

func (m *mockFollowRepo) GetFollowStatus(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*models.Follow, error) {
	if m.GetFollowStatusFunc != nil {
		return m.GetFollowStatusFunc(ctx, followerID, followingID)
	}
	return &models.Follow{FollowerID: followerID, FollowingID: followingID}, nil
}

func (m *mockFollowRepo) GetFollowers(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Follow, error) {
	if m.GetFollowersFunc != nil {
		return m.GetFollowersFunc(ctx, userID, limit, offset)
	}
	return []*models.Follow{}, nil
}

func (m *mockFollowRepo) GetFollowing(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Follow, error) {
	if m.GetFollowingFunc != nil {
		return m.GetFollowingFunc(ctx, userID, limit, offset)
	}
	return []*models.Follow{}, nil
}

// --- Tests ---

func TestFollowUser_Success(t *testing.T) {
	h := NewFollowHandler(&mockFollowRepo{})

	req := httptest.NewRequest("POST", "/users/00000000-0000-0000-0000-000000000001/follow", nil)
	// mock path parsing manually in handler or trust URL parse.
	// We used Query param 'id' fallback, so putting it in query just in case for test simplicity if mux is absent.
	// But handler logic tries to split path.
	// To make `httptest.NewRequest` work with path parsing logic in handler, we need ensuring URL matches expected structure.
	// req.URL.Path is set to the checked value.

	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.FollowUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestUnfollowUser_Success(t *testing.T) {
	h := NewFollowHandler(&mockFollowRepo{})

	req := httptest.NewRequest("DELETE", "/users/00000000-0000-0000-0000-000000000001/follow", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnfollowUser(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestUnfollowUser_NotFound(t *testing.T) {
	mockRepo := &mockFollowRepo{
		UnfollowFunc: func(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
			return repository.ErrFollowNotFound
		},
	}
	h := NewFollowHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/users/00000000-0000-0000-0000-000000000001/follow", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnfollowUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

func TestGetFollowers_Success(t *testing.T) {
	h := NewFollowHandler(&mockFollowRepo{})

	req := httptest.NewRequest("GET", "/users/00000000-0000-0000-0000-000000000001/followers", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetFollowers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestGetFollowing_Success(t *testing.T) {
	h := NewFollowHandler(&mockFollowRepo{})

	req := httptest.NewRequest("GET", "/users/00000000-0000-0000-0000-000000000001/following", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetFollowing(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

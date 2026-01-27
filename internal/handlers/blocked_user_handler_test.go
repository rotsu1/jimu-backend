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

type mockBlockedUserRepo struct {
	BlockFunc           func(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) (*models.BlockedUser, error)
	UnblockFunc         func(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) error
	GetBlockedUsersFunc func(ctx context.Context, blockerID uuid.UUID) ([]*models.BlockedUser, error)
}

func (m *mockBlockedUserRepo) Block(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) (*models.BlockedUser, error) {
	if m.BlockFunc != nil {
		return m.BlockFunc(ctx, blockerID, blockedID)
	}
	return &models.BlockedUser{}, nil
}

func (m *mockBlockedUserRepo) Unblock(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) error {
	if m.UnblockFunc != nil {
		return m.UnblockFunc(ctx, blockerID, blockedID)
	}
	return nil
}

func (m *mockBlockedUserRepo) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]*models.BlockedUser, error) {
	if m.GetBlockedUsersFunc != nil {
		return m.GetBlockedUsersFunc(ctx, blockerID)
	}
	return []*models.BlockedUser{}, nil
}

// --- Tests ---

func TestBlockUser_Success(t *testing.T) {
	h := NewBlockedUserHandler(&mockBlockedUserRepo{})

	body := `{"blocked_id": "00000000-0000-0000-0000-000000000001"}`
	req := httptest.NewRequest("POST", "/blocked-users", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.BlockUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}
}

func TestBlockUser_InvalidInput(t *testing.T) {
	h := NewBlockedUserHandler(&mockBlockedUserRepo{})

	body := `{"blocked_id": "invalid"}`
	req := httptest.NewRequest("POST", "/blocked-users", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.BlockUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestBlockUser_ReferenceViolation(t *testing.T) {
	mockRepo := &mockBlockedUserRepo{
		BlockFunc: func(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) (*models.BlockedUser, error) {
			return nil, repository.ErrReferenceViolation
		},
	}
	h := NewBlockedUserHandler(mockRepo)

	body := `{"blocked_id": "00000000-0000-0000-0000-000000000001"}`
	req := httptest.NewRequest("POST", "/blocked-users", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.BlockUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

func TestUnblockUser_Success(t *testing.T) {
	h := NewBlockedUserHandler(&mockBlockedUserRepo{})

	req := httptest.NewRequest("DELETE", "/blocked-users?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnblockUser(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestUnblockUser_NotFound(t *testing.T) {
	mockRepo := &mockBlockedUserRepo{
		UnblockFunc: func(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) error {
			return repository.ErrBlockedUserNotFound
		},
	}
	h := NewBlockedUserHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/blocked-users?id=00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UnblockUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

func TestGetBlockedUsers_Success(t *testing.T) {
	h := NewBlockedUserHandler(&mockBlockedUserRepo{})

	req := httptest.NewRequest("GET", "/blocked-users", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetBlockedUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

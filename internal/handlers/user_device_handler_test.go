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

type mockUserDeviceRepo struct {
	UpsertUserDeviceFunc          func(ctx context.Context, userID uuid.UUID, fcmToken string, deviceType string, viewerID uuid.UUID) (*models.UserDevice, error)
	GetUserDeviceByIDFunc         func(ctx context.Context, deviceID uuid.UUID, userID uuid.UUID) (*models.UserDevice, error)
	GetUserDevicesByUserIDFunc    func(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID) ([]*models.UserDevice, error)
	GetUserDeviceByFCMTokenFunc   func(ctx context.Context, fcmToken string, userID uuid.UUID) (*models.UserDevice, error)
	DeleteUserDeviceByIDFunc      func(ctx context.Context, deviceID uuid.UUID, userID uuid.UUID) error
	DeleteUserDevicesByUserIDFunc func(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID) error
}

func (m *mockUserDeviceRepo) UpsertUserDevice(ctx context.Context, userID uuid.UUID, fcmToken string, deviceType string, viewerID uuid.UUID) (*models.UserDevice, error) {
	if m.UpsertUserDeviceFunc != nil {
		return m.UpsertUserDeviceFunc(ctx, userID, fcmToken, deviceType, viewerID)
	}
	return &models.UserDevice{UserID: userID, FCMToken: fcmToken}, nil
}

func (m *mockUserDeviceRepo) GetUserDeviceByID(ctx context.Context, deviceID uuid.UUID, userID uuid.UUID) (*models.UserDevice, error) {
	if m.GetUserDeviceByIDFunc != nil {
		return m.GetUserDeviceByIDFunc(ctx, deviceID, userID)
	}
	return &models.UserDevice{ID: deviceID}, nil
}

func (m *mockUserDeviceRepo) GetUserDevicesByUserID(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID) ([]*models.UserDevice, error) {
	if m.GetUserDevicesByUserIDFunc != nil {
		return m.GetUserDevicesByUserIDFunc(ctx, targetID, viewerID)
	}
	return []*models.UserDevice{}, nil
}

func (m *mockUserDeviceRepo) GetUserDeviceByFCMToken(ctx context.Context, fcmToken string, userID uuid.UUID) (*models.UserDevice, error) {
	if m.GetUserDeviceByFCMTokenFunc != nil {
		return m.GetUserDeviceByFCMTokenFunc(ctx, fcmToken, userID)
	}
	return nil, nil
}

func (m *mockUserDeviceRepo) DeleteUserDeviceByID(ctx context.Context, deviceID uuid.UUID, userID uuid.UUID) error {
	if m.DeleteUserDeviceByIDFunc != nil {
		return m.DeleteUserDeviceByIDFunc(ctx, deviceID, userID)
	}
	return nil
}

func (m *mockUserDeviceRepo) DeleteUserDevicesByUserID(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID) error {
	if m.DeleteUserDevicesByUserIDFunc != nil {
		return m.DeleteUserDevicesByUserIDFunc(ctx, targetID, viewerID)
	}
	return nil
}

// --- Tests ---

func TestRegisterDevice_Success(t *testing.T) {
	h := NewUserDeviceHandler(&mockUserDeviceRepo{})

	body := `{"fcm_token": "token123", "device_type": "ios"}`
	req := httptest.NewRequest("POST", "/user-devices", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.RegisterDevice(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestListDevices_Success(t *testing.T) {
	h := NewUserDeviceHandler(&mockUserDeviceRepo{})

	req := httptest.NewRequest("GET", "/user-devices", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.ListDevices(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestDeleteDevice_Success(t *testing.T) {
	h := NewUserDeviceHandler(&mockUserDeviceRepo{})

	req := httptest.NewRequest("DELETE", "/user-devices/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteDevice(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestDeleteDevice_NotFound(t *testing.T) {
	mockRepo := &mockUserDeviceRepo{
		DeleteUserDeviceByIDFunc: func(ctx context.Context, deviceID uuid.UUID, userID uuid.UUID) error {
			return repository.ErrUserDeviceNotFound
		},
	}
	h := NewUserDeviceHandler(mockRepo)

	req := httptest.NewRequest("DELETE", "/user-devices/00000000-0000-0000-0000-000000000001", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.DeleteDevice(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

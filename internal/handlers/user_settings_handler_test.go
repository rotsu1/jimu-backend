package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/handlers/testutils"
	"github.com/rotsu1/jimu-backend/internal/models"
)

// --- Mocks ---

type mockUserSettingsRepo struct {
	GetUserSettingsByIDFunc func(ctx context.Context, id uuid.UUID) (*models.UserSetting, error)
	UpdateUserSettingsFunc  func(ctx context.Context, id string, updates models.UpdateUserSettingsRequest) error
}

func (m *mockUserSettingsRepo) GetUserSettingsByID(ctx context.Context, id uuid.UUID) (*models.UserSetting, error) {
	if m.GetUserSettingsByIDFunc != nil {
		return m.GetUserSettingsByIDFunc(ctx, id)
	}
	return &models.UserSetting{UserID: id}, nil
}

func (m *mockUserSettingsRepo) UpdateUserSettings(ctx context.Context, id string, updates models.UpdateUserSettingsRequest) error {
	if m.UpdateUserSettingsFunc != nil {
		return m.UpdateUserSettingsFunc(ctx, id, updates)
	}
	return nil
}

// --- Tests ---

func TestGetMySettings_Success(t *testing.T) {
	h := NewUserSettingsHandler(&mockUserSettingsRepo{})

	req := httptest.NewRequest("GET", "/settings", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetMySettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestUpdateMySettings_Success(t *testing.T) {
	h := NewUserSettingsHandler(&mockUserSettingsRepo{})

	body := `{"sound_enabled": true}`
	req := httptest.NewRequest("PUT", "/settings", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpdateMySettings(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", rr.Code)
	}
}

func TestUpdateMySettings_Error(t *testing.T) {
	mockRepo := &mockUserSettingsRepo{
		UpdateUserSettingsFunc: func(ctx context.Context, id string, updates models.UpdateUserSettingsRequest) error {
			return errors.New("db error")
		},
	}
	h := NewUserSettingsHandler(mockRepo)

	body := `{"sound_enabled": true}`
	req := httptest.NewRequest("PUT", "/settings", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpdateMySettings(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rr.Code)
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/auth"
	"github.com/rotsu1/jimu-backend/internal/models"
	"google.golang.org/api/idtoken"
)

// --- Mocks ---

type mockUserRepo struct {
	UpsertGoogleUserFunc func(ctx context.Context, googleID, email string) (*models.Profile, error)
}

func (m *mockUserRepo) UpsertGoogleUser(ctx context.Context, googleID, email string) (*models.Profile, error) {
	if m.UpsertGoogleUserFunc != nil {
		return m.UpsertGoogleUserFunc(ctx, googleID, email)
	}
	return &models.Profile{ID: uuid.New()}, nil
}

type mockSessionRepo struct {
	CreateSessionFunc            func(ctx context.Context, userID uuid.UUID, token string, agent *string, ip *netip.Addr, exp time.Time) (*models.UserSession, error)
	GetSessionByRefreshTokenFunc func(ctx context.Context, token string) (*models.UserSession, error)
	RevokeSessionFunc            func(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) error
	RevokeAllSessionsForUserFunc func(ctx context.Context, targetUserID uuid.UUID, viewerID uuid.UUID) error
}

func (m *mockSessionRepo) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	token string,
	agent *string,
	ip *netip.Addr,
	exp time.Time,
) (*models.UserSession, error) {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, userID, token, agent, ip, exp)
	}
	return &models.UserSession{ID: uuid.New(), UserID: userID}, nil
}

func (m *mockSessionRepo) GetSessionByRefreshToken(
	ctx context.Context,
	token string,
) (*models.UserSession, error) {
	if m.GetSessionByRefreshTokenFunc != nil {
		return m.GetSessionByRefreshTokenFunc(ctx, token)
	}
	return &models.UserSession{ID: uuid.New(), UserID: uuid.New()}, nil
}

func (m *mockSessionRepo) RevokeSession(
	ctx context.Context,
	id uuid.UUID,
	viewerID uuid.UUID,
) error {
	if m.RevokeSessionFunc != nil {
		return m.RevokeSessionFunc(ctx, id, viewerID)
	}
	return nil
}

func (m *mockSessionRepo) RevokeAllSessionsForUser(
	ctx context.Context,
	targetUserID uuid.UUID,
	viewerID uuid.UUID,
) error {
	if m.RevokeAllSessionsForUserFunc != nil {
		return m.RevokeAllSessionsForUserFunc(ctx, targetUserID, viewerID)
	}
	return nil
}

type mockValidator struct {
	ValidateFunc func(ctx context.Context, token string, aud string) (*idtoken.Payload, error)
}

func (m *mockValidator) Validate(
	ctx context.Context,
	token string,
	aud string,
) (*idtoken.Payload, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx, token, aud)
	}
	return &idtoken.Payload{
		Subject: "mock-google-id",
		Claims:  map[string]interface{}{"email": "test@jimu.com"},
	}, nil
}

// --- Tests ---

func TestGoogleLogin_Success(t *testing.T) {
	mockRepo := &mockUserRepo{}
	mockSessionRepo := &mockSessionRepo{}
	mockValidator := &mockValidator{}
	h := NewAuthHandler(mockRepo, mockSessionRepo, mockValidator)

	body := `{"id_token": "valid-token"}`
	req := httptest.NewRequest("POST", "/auth/google", strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()

	h.GoogleLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
}

func TestGoogleLogin_InvalidToken(t *testing.T) {
	mockValidator := &mockValidator{
		ValidateFunc: func(ctx context.Context, token string, aud string) (*idtoken.Payload, error) {
			return nil, errors.New("invalid token")
		},
	}
	h := NewAuthHandler(&mockUserRepo{}, &mockSessionRepo{}, mockValidator)

	body := `{"id_token": "invalid-token"}`
	req := httptest.NewRequest("POST", "/auth/google", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.GoogleLogin(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}
}

func TestGoogleLogin_DatabaseDown(t *testing.T) {
	mockRepo := &mockUserRepo{
		UpsertGoogleUserFunc: func(ctx context.Context, googleID, email string) (*models.Profile, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewAuthHandler(mockRepo, &mockSessionRepo{}, &mockValidator{})

	body := `{"id_token": "valid-token"}`
	req := httptest.NewRequest("POST", "/auth/google", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.GoogleLogin(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rr.Code)
	}
}

func TestGoogleLogin_SessionSaveFail(t *testing.T) {
	mockSessionRepo := &mockSessionRepo{
		CreateSessionFunc: func(ctx context.Context, userID uuid.UUID, token string, agent *string, ip *netip.Addr, exp time.Time) (*models.UserSession, error) {
			return nil, errors.New("session save error")
		},
	}
	h := NewAuthHandler(&mockUserRepo{}, mockSessionRepo, &mockValidator{})

	body := `{"id_token": "valid-token"}`
	req := httptest.NewRequest("POST", "/auth/google", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.GoogleLogin(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rr.Code)
	}
}

func TestGoogleLogin_MissingBody(t *testing.T) {
	h := NewAuthHandler(&mockUserRepo{}, &mockSessionRepo{}, &mockValidator{})

	req := httptest.NewRequest("POST", "/auth/google", strings.NewReader(""))
	rr := httptest.NewRecorder()

	h.GoogleLogin(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rr.Code)
	}
}

// --- Refresh Token Tests ---

func TestRefreshToken_ExpiredInvalid(t *testing.T) {
	mockSessionRepo := &mockSessionRepo{
		GetSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.UserSession, error) {
			return nil, errors.New("not found")
		},
	}
	h := NewAuthHandler(&mockUserRepo{}, mockSessionRepo, &mockValidator{})

	body := `{"refresh_token": "expired-token"}`
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.RefreshToken(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}
}

func TestRefreshToken_RotationSuccess(t *testing.T) {
	revokeCalled := false
	createCalled := false
	oldSessionID := uuid.New()
	userID := uuid.New()

	mockSessionRepo := &mockSessionRepo{
		GetSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.UserSession, error) {
			return &models.UserSession{ID: oldSessionID, UserID: userID}, nil
		},
		RevokeSessionFunc: func(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) error {
			if id == oldSessionID && viewerID == userID {
				revokeCalled = true
			}
			return nil
		},
		CreateSessionFunc: func(ctx context.Context, uid uuid.UUID, token string, agent *string, ip *netip.Addr, exp time.Time) (*models.UserSession, error) {
			if uid == userID {
				createCalled = true
			}
			return &models.UserSession{ID: uuid.New(), UserID: uid}, nil
		},
	}
	h := NewAuthHandler(&mockUserRepo{}, mockSessionRepo, &mockValidator{})
	// Mock generating token pair by setting secret env
	os.Setenv("JIMU_SECRET", "test-secret")
	defer os.Unsetenv("JIMU_SECRET")

	body := `{"refresh_token": "valid-token"}`
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()

	h.RefreshToken(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if !revokeCalled {
		t.Error("old session was not revoked")
	}
	if !createCalled {
		t.Error("new session was not created")
	}

	var resp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["access_token"] == "" || resp["refresh_token"] == "" {
		t.Error("expected access_token and refresh_token")
	}
}

func TestRefreshToken_WonkyIP(t *testing.T) {
	mockSessionRepo := &mockSessionRepo{
		GetSessionByRefreshTokenFunc: func(ctx context.Context, token string) (*models.UserSession, error) {
			return &models.UserSession{ID: uuid.New(), UserID: uuid.New()}, nil
		},
		CreateSessionFunc: func(ctx context.Context, uid uuid.UUID, token string, agent *string, ip *netip.Addr, exp time.Time) (*models.UserSession, error) {
			return &models.UserSession{ID: uuid.New(), UserID: uid}, nil
		},
	}
	os.Setenv("JIMU_SECRET", "test-secret")
	defer os.Unsetenv("JIMU_SECRET")

	h := NewAuthHandler(&mockUserRepo{}, mockSessionRepo, &mockValidator{})

	body := `{"refresh_token": "valid-token"}`
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.RemoteAddr = "wonky-ip"
	rr := httptest.NewRecorder()

	h.RefreshToken(rr, req)

	// NOTE: The current implementation returns 500 on wonky IP.
	// If this test fails with 500, we should fix the handler implementation.
	// But based on requirement "this should still result in a 200 OK with a nil IP",
	// I assume I need to assert 200 OK.
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK for wonky IP, got %d", rr.Code)
	}
}

// --- Logout Tests ---

func TestLogout_MissingHeader(t *testing.T) {
	h := NewAuthHandler(&mockUserRepo{}, &mockSessionRepo{}, &mockValidator{})

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}
}

func TestLogout_MalformedHeader(t *testing.T) {
	h := NewAuthHandler(&mockUserRepo{}, &mockSessionRepo{}, &mockValidator{})

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Basic 123")
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}
}

func TestLogout_TokenVerificationFail(t *testing.T) {
	h := NewAuthHandler(&mockUserRepo{}, &mockSessionRepo{}, &mockValidator{})

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}
}

func TestLogout_RevokeFail(t *testing.T) {
	secret := "test-secret"
	os.Setenv("JIMU_SECRET", secret)
	defer os.Unsetenv("JIMU_SECRET")

	uid := uuid.New()
	token, _, _ := auth.GenerateTokenPair(uid.String(), secret)

	mockSessionRepo := &mockSessionRepo{
		RevokeAllSessionsForUserFunc: func(ctx context.Context, targetUserID uuid.UUID, viewerID uuid.UUID) error {
			return errors.New("db error")
		},
	}
	h := NewAuthHandler(&mockUserRepo{}, mockSessionRepo, &mockValidator{})

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	h.Logout(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rr.Code)
	}
}

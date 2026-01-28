package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/testutil"
)

// =============================================================================
// AuthHandler Profile Tests
// =============================================================================

// TestIntegration_Auth_GetMyProfile tests getting the current user's profile.
func TestIntegration_Auth_GetMyProfile(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "profile-test-user")
	token := testutil.CreateTestToken(user.ID)

	// Act - GET /auth/profile
	req := httptest.NewRequest("GET", "/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("GET /auth/profile: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "profile-test-user") {
		t.Errorf("Expected response to contain username, got: %s", rr.Body.String())
	}
}

// TestIntegration_Auth_UpdateMyProfile tests updating the current user's profile.
func TestIntegration_Auth_UpdateMyProfile(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "update-profile-user")
	token := testutil.CreateTestToken(user.ID)

	// Act - PUT /auth/profile
	payload := `{"username": "updated-username", "bio": "Test bio"}`
	req := httptest.NewRequest("PUT", "/auth/profile", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// Assert
	if rr.Code != http.StatusNoContent {
		t.Errorf("PUT /auth/profile: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify database
	var username string
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT username FROM profiles WHERE id = $1",
		user.ID,
	).Scan(&username)
	if err != nil {
		t.Fatalf("Failed to query profile: %v", err)
	}
	if username != "updated-username" {
		t.Errorf("Expected username 'updated-username', got '%s'", username)
	}
}

// TestIntegration_Auth_GetOtherProfile tests viewing another user's profile.
func TestIntegration_Auth_GetOtherProfile(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	userA := srv.SeedUser(t, "viewer-user")
	userB := srv.SeedUser(t, "target-profile-user")
	tokenA := testutil.CreateTestToken(userA.ID)

	// Act - GET /auth/profile/{id}
	req := httptest.NewRequest("GET", "/auth/profile/"+userB.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("GET /auth/profile/{id}: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "target-profile-user") {
		t.Errorf("Expected response to contain target username, got: %s", rr.Body.String())
	}
}

// TestIntegration_Auth_GetOtherProfile_NotFound tests viewing a non-existent profile.
func TestIntegration_Auth_GetOtherProfile_NotFound(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "viewer-notfound-user")
	token := testutil.CreateTestToken(user.ID)

	fakeID := uuid.New()
	req := httptest.NewRequest("GET", "/auth/profile/"+fakeID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent profile, got %d: %s", rr.Code, rr.Body.String())
	}
}

// =============================================================================
// AuthHandler RefreshToken Tests
// =============================================================================

// TestIntegration_Auth_RefreshToken tests the token refresh flow.
func TestIntegration_Auth_RefreshToken(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "refresh-token-user")

	// Seed a valid user_session with a known refresh token
	refreshToken := "test-refresh-token-" + uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := srv.DB.Exec(
		context.Background(),
		`INSERT INTO user_sessions (user_id, refresh_token, expires_at)
		 VALUES ($1, $2, $3)`,
		user.ID, refreshToken, expiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to seed user session: %v", err)
	}

	// Act - POST /auth/refresh
	payload := `{"refresh_token": "` + refreshToken + `"}`
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("POST /auth/refresh: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify response contains new tokens
	body := rr.Body.String()
	if !strings.Contains(body, "access_token") {
		t.Errorf("Expected response to contain 'access_token', got: %s", body)
	}
	if !strings.Contains(body, "refresh_token") {
		t.Errorf("Expected response to contain 'refresh_token', got: %s", body)
	}

	// Verify old session was revoked (or replaced)
	var isRevoked bool
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT is_revoked FROM user_sessions WHERE refresh_token = $1",
		refreshToken,
	).Scan(&isRevoked)
	if err != nil {
		t.Fatalf("Failed to query revoked session: %v", err)
	}
	if !isRevoked {
		t.Errorf("Expected old session to be revoked")
	}
}

// TestIntegration_Auth_RefreshToken_Invalid tests refresh with an invalid token.
func TestIntegration_Auth_RefreshToken_Invalid(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	// Act - POST /auth/refresh with invalid token
	payload := `{"refresh_token": "invalid-token-that-does-not-exist"}`
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// Assert
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid refresh token, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestIntegration_Auth_Unauthorized tests accessing profile without token.
func TestIntegration_Auth_Unauthorized(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for unauthenticated request, got %d", rr.Code)
	}
}

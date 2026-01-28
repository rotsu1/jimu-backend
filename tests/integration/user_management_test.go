package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/testutil"
)

// =============================================================================
// UserSettingsHandler Tests
// =============================================================================

// TestIntegration_UserSettings_GetAndUpdate tests getting and updating user settings.
func TestIntegration_UserSettings_GetAndUpdate(t *testing.T) {
	// 1. Arrange
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "settings-test-user")
	token := testutil.CreateTestToken(user.ID)

	// Seed initial user_settings (they may be auto-created on profile creation,
	// but let's ensure they exist)
	_, err := srv.DB.Exec(
		context.Background(),
		`INSERT INTO user_settings (user_id)
		 VALUES ($1)
		 ON CONFLICT (user_id) DO NOTHING`,
		user.ID,
	)
	if err != nil {
		t.Fatalf("Failed to seed user settings: %v", err)
	}

	// 2. Act - GET /user-settings
	req := httptest.NewRequest("GET", "/user-settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// 3. Assert
	if rr.Code != http.StatusOK {
		t.Errorf("GET /user-settings: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Act - PUT /user-settings
	updatePayload := `{"notify_new_follower": false}`
	req = httptest.NewRequest("PUT", "/user-settings", strings.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// 5. Assert update response
	if rr.Code != http.StatusNoContent {
		t.Errorf("PUT /user-settings: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify database state
	var notifyNewFollower bool
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT notify_new_follower FROM user_settings WHERE user_id = $1",
		user.ID,
	).Scan(&notifyNewFollower)
	if err != nil {
		t.Fatalf("Failed to query user settings: %v", err)
	}
	if notifyNewFollower {
		t.Errorf("Expected notify_new_follower to be false, got true")
	}
}

// TestIntegration_UserSettings_Unauthorized tests that unauthenticated requests are rejected.
func TestIntegration_UserSettings_Unauthorized(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	req := httptest.NewRequest("GET", "/user-settings", nil)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rr.Code)
	}
}

// =============================================================================
// UserDeviceHandler Tests
// =============================================================================

// TestIntegration_UserDevice_Lifecycle tests registering, listing, and deleting devices.
func TestIntegration_UserDevice_Lifecycle(t *testing.T) {
	// 1. Arrange
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "device-test-user")
	token := testutil.CreateTestToken(user.ID)

	// 2. Act - POST /user-devices (register device)
	payload := `{"fcm_token": "test-fcm-token-12345", "device_type": "ios"}`
	req := httptest.NewRequest("POST", "/user-devices", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// 3. Assert
	if rr.Code != http.StatusOK {
		t.Errorf("POST /user-devices: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Verify database
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM user_devices WHERE user_id = $1",
		user.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query user devices: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 device, got %d", count)
	}

	// 5. Act - GET /user-devices (list devices)
	req = httptest.NewRequest("GET", "/user-devices", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /user-devices: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "test-fcm-token-12345") {
		t.Errorf("Expected response to contain fcm_token, got: %s", rr.Body.String())
	}

	// 6. Get device ID for deletion
	var deviceID uuid.UUID
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM user_devices WHERE user_id = $1 LIMIT 1",
		user.ID,
	).Scan(&deviceID)
	if err != nil {
		t.Fatalf("Failed to get device ID: %v", err)
	}

	// 7. Act - DELETE /user-devices?id={deviceID}
	req = httptest.NewRequest("DELETE", "/user-devices?id="+deviceID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /user-devices: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 8. Verify deletion
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM user_devices WHERE user_id = $1",
		user.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query user devices after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 devices after deletion, got %d", count)
	}
}

// TestIntegration_UserDevice_RegisterWithoutToken tests that FCM token is required.
func TestIntegration_UserDevice_RegisterWithoutToken(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "device-token-test-user")
	token := testutil.CreateTestToken(user.ID)

	// Missing fcm_token
	payload := `{"device_type": "android"}`
	req := httptest.NewRequest("POST", "/user-devices", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing FCM token, got %d", rr.Code)
	}
}

// TestIntegration_UserDevice_DeleteNonexistent tests deleting a device that doesn't exist.
func TestIntegration_UserDevice_DeleteNonexistent(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "device-delete-test-user")
	token := testutil.CreateTestToken(user.ID)

	fakeID := uuid.New()
	req := httptest.NewRequest("DELETE", "/user-devices?id="+fakeID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for nonexistent device, got %d: %s", rr.Code, rr.Body.String())
	}
}

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rotsu1/jimu-backend/internal/testutil"
)

var testServer *testutil.TestServer

// TestMain handles global test database lifecycle.
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup: close database pool if initialized
	if testServer != nil && testServer.DB != nil {
		testServer.DB.Close()
	}

	os.Exit(code)
}

// TestIntegration_WorkoutLifecycle tests the complete workout creation flow
// using real HTTP requests through the router with a real database.
func TestIntegration_WorkoutLifecycle(t *testing.T) {
	// 1. Setup - Create test server with real dependencies
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	// 2. Seed - Create a test user
	user := srv.SeedUser(t, "integration-test-user")

	// 3. Generate auth token
	token := testutil.CreateTestToken(user.ID)

	// 4. Act - Create a workout via HTTP POST
	payload := `{
		"name": "Morning Chest Day",
		"comment": "Integration test workout",
		"started_at": "2026-01-28T08:00:00Z",
		"ended_at": "2026-01-28T09:30:00Z",
		"duration_seconds": 5400
	}`

	req := httptest.NewRequest("POST", "/workouts", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Execute request through the router
	srv.Router.ServeHTTP(rr, req)

	// 5. Assert - Check HTTP response
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify - Query database directly to confirm workout exists
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM workouts WHERE user_id = $1",
		user.ID,
	).Scan(&count)

	if err != nil {
		t.Fatalf("Failed to query workouts: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 workout in database, got %d", count)
	}
}

// TestIntegration_WorkoutList tests listing workouts for the authenticated user.
func TestIntegration_WorkoutList(t *testing.T) {
	// 1. Setup
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	// 2. Seed user
	user := srv.SeedUser(t, "list-test-user")
	token := testutil.CreateTestToken(user.ID)

	// 3. Seed workout directly in DB
	_, err := srv.DB.Exec(
		context.Background(),
		`INSERT INTO workouts (user_id, name, started_at, ended_at, duration_seconds) 
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID,
		"Seeded Workout",
		time.Now().Add(-1*time.Hour),
		time.Now(),
		3600,
	)
	if err != nil {
		t.Fatalf("Failed to seed workout: %v", err)
	}

	// 4. Act - List workouts via HTTP GET
	req := httptest.NewRequest("GET", "/workouts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// 5. Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	// Check response contains expected workout (simple substring check)
	if !strings.Contains(rr.Body.String(), "Seeded Workout") {
		t.Errorf("expected response to contain 'Seeded Workout', got: %s", rr.Body.String())
	}
}

// TestIntegration_WorkoutUnauthorized tests that requests without auth are rejected.
func TestIntegration_WorkoutUnauthorized(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	// Act - Try to list workouts without auth header
	req := httptest.NewRequest("GET", "/workouts", nil)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	// Assert - Should get 401 Unauthorized
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", rr.Code)
	}
}

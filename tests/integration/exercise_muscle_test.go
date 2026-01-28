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
// ExerciseHandler Tests
// =============================================================================

// TestIntegration_Exercise_Lifecycle tests the full exercise CRUD lifecycle.
func TestIntegration_Exercise_Lifecycle(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "exercise-user")
	token := testutil.CreateTestToken(user.ID)

	// 1. Act - POST /exercises
	payload := `{"name": "Deadlift", "suggested_rest_seconds": 180}`
	req := httptest.NewRequest("POST", "/exercises", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /exercises: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database and get exercise ID
	var exerciseID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM exercises WHERE user_id = $1 AND name = $2",
		user.ID, "Deadlift",
	).Scan(&exerciseID)
	if err != nil {
		t.Fatalf("Failed to query exercise: %v", err)
	}

	// 3. Act - GET /exercises (list)
	req = httptest.NewRequest("GET", "/exercises", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /exercises: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "Deadlift") {
		t.Errorf("Expected response to contain 'Deadlift', got: %s", rr.Body.String())
	}

	// 4. Act - GET /exercises/{id}
	req = httptest.NewRequest("GET", "/exercises/"+exerciseID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /exercises/{id}: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Act - PUT /exercises/{id}
	updatePayload := `{"name": "Romanian Deadlift"}`
	req = httptest.NewRequest("PUT", "/exercises/"+exerciseID.String(), strings.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("PUT /exercises/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify update
	var name string
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT name FROM exercises WHERE id = $1",
		exerciseID,
	).Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query updated exercise: %v", err)
	}
	if name != "Romanian Deadlift" {
		t.Errorf("Expected name 'Romanian Deadlift', got '%s'", name)
	}

	// 7. Act - DELETE /exercises/{id}
	req = httptest.NewRequest("DELETE", "/exercises/"+exerciseID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /exercises/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 8. Verify deletion
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM exercises WHERE id = $1",
		exerciseID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query exercises after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 exercises after deletion, got %d", count)
	}
}

// =============================================================================
// MuscleHandler Tests
// =============================================================================

// TestIntegration_Muscle_Lifecycle tests the muscle CRUD lifecycle.
func TestIntegration_Muscle_Lifecycle(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "muscle-user")
	token := testutil.CreateTestToken(user.ID)

	// Seed as sys_admin to allow muscle creation (if required)
	_, _ = srv.DB.Exec(
		context.Background(),
		"INSERT INTO sys_admins (user_id) VALUES ($1) ON CONFLICT DO NOTHING",
		user.ID,
	)

	// 1. Act - POST /muscles
	payload := `{"name": "Quadriceps"}`
	req := httptest.NewRequest("POST", "/muscles", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /muscles: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Get muscle ID
	var muscleID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM muscles WHERE name = $1",
		"Quadriceps",
	).Scan(&muscleID)
	if err != nil {
		t.Fatalf("Failed to query muscle: %v", err)
	}

	// 3. Act - GET /muscles (list)
	req = httptest.NewRequest("GET", "/muscles", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /muscles: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "Quadriceps") {
		t.Errorf("Expected response to contain 'Quadriceps', got: %s", rr.Body.String())
	}

	// 4. Act - GET /muscles/{id}
	req = httptest.NewRequest("GET", "/muscles/"+muscleID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /muscles/{id}: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Act - DELETE /muscles/{id}
	req = httptest.NewRequest("DELETE", "/muscles/"+muscleID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /muscles/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify deletion
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM muscles WHERE id = $1",
		muscleID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query muscles after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 muscles after deletion, got %d", count)
	}
}

// =============================================================================
// ExerciseTargetMuscleHandler Tests
// =============================================================================

// TestIntegration_ExerciseTargetMuscle_AddAndRemove tests adding/removing target muscles.
func TestIntegration_ExerciseTargetMuscle_AddAndRemove(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "target-muscle-user")
	token := testutil.CreateTestToken(user.ID)

	// Seed exercise
	var exerciseID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		"INSERT INTO exercises (user_id, name) VALUES ($1, $2) RETURNING id",
		user.ID, "Leg Press",
	).Scan(&exerciseID)
	if err != nil {
		t.Fatalf("Failed to seed exercise: %v", err)
	}

	// Seed muscle
	var muscleID uuid.UUID
	err = srv.DB.QueryRow(
		context.Background(),
		"INSERT INTO muscles (name) VALUES ($1) RETURNING id",
		"Hamstrings",
	).Scan(&muscleID)
	if err != nil {
		t.Fatalf("Failed to seed muscle: %v", err)
	}

	// 1. Act - POST /exercises/{id}/muscles
	payload := `{"muscle_id": "` + muscleID.String() + `"}`
	req := httptest.NewRequest("POST", "/exercises/"+exerciseID.String()+"/muscles", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /exercises/{id}/muscles: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM exercise_target_muscles WHERE exercise_id = $1 AND muscle_id = $2",
		exerciseID, muscleID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query exercise_target_muscles: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 exercise_target_muscle, got %d", count)
	}

	// 3. Act - DELETE /exercises/{id}/muscles/{muscleId}
	req = httptest.NewRequest("DELETE", "/exercises/"+exerciseID.String()+"/muscles/"+muscleID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /exercises/{id}/muscles/{muscleId}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Verify deletion
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM exercise_target_muscles WHERE exercise_id = $1 AND muscle_id = $2",
		exerciseID, muscleID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query exercise_target_muscles after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 exercise_target_muscles after deletion, got %d", count)
	}
}

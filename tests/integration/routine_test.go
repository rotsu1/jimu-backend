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
// Helper: seedExerciseForRoutine creates an exercise for routine tests
// =============================================================================

func seedExerciseForRoutine(t *testing.T, srv *testutil.TestServer, userID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	var exerciseID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO exercises (user_id, name) VALUES ($1, $2) RETURNING id`,
		userID, name,
	).Scan(&exerciseID)
	if err != nil {
		t.Fatalf("Failed to seed exercise: %v", err)
	}
	return exerciseID
}

// seedRoutine creates a routine for the given user
func seedRoutine(t *testing.T, srv *testutil.TestServer, userID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	var routineID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO routines (user_id, name) VALUES ($1, $2) RETURNING id`,
		userID, name,
	).Scan(&routineID)
	if err != nil {
		t.Fatalf("Failed to seed routine: %v", err)
	}
	return routineID
}

// seedRoutineExercise creates a routine exercise
func seedRoutineExercise(t *testing.T, srv *testutil.TestServer, routineID, exerciseID uuid.UUID) uuid.UUID {
	t.Helper()
	var reID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO routine_exercises (routine_id, exercise_id, order_index)
		 VALUES ($1, $2, 1) RETURNING id`,
		routineID, exerciseID,
	).Scan(&reID)
	if err != nil {
		t.Fatalf("Failed to seed routine exercise: %v", err)
	}
	return reID
}

// =============================================================================
// RoutineHandler Tests
// =============================================================================

// TestIntegration_Routine_Lifecycle tests the full routine CRUD lifecycle.
func TestIntegration_Routine_Lifecycle(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "routine-user")
	token := testutil.CreateTestToken(user.ID)

	// 1. Act - POST /routines
	payload := `{"name": "Push Day", "description": "Chest, shoulders, triceps"}`
	req := httptest.NewRequest("POST", "/routines", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /routines: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database and get routine ID
	var routineID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM routines WHERE user_id = $1 AND name = $2",
		user.ID, "Push Day",
	).Scan(&routineID)
	if err != nil {
		t.Fatalf("Failed to query routine: %v", err)
	}

	// 3. Act - GET /routines (list)
	req = httptest.NewRequest("GET", "/routines", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /routines: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "Push Day") {
		t.Errorf("Expected response to contain 'Push Day', got: %s", rr.Body.String())
	}

	// 4. Act - GET /routines/{id}
	req = httptest.NewRequest("GET", "/routines/"+routineID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /routines/{id}: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Act - PUT /routines/{id}
	updatePayload := `{"name": "Pull Day"}`
	req = httptest.NewRequest("PUT", "/routines/"+routineID.String(), strings.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("PUT /routines/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify update
	var name string
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT name FROM routines WHERE id = $1",
		routineID,
	).Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query updated routine: %v", err)
	}
	if name != "Pull Day" {
		t.Errorf("Expected name 'Pull Day', got '%s'", name)
	}

	// 7. Act - DELETE /routines/{id}
	req = httptest.NewRequest("DELETE", "/routines/"+routineID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /routines/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 8. Verify deletion
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM routines WHERE id = $1",
		routineID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query routines after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 routines after deletion, got %d", count)
	}
}

// TestIntegration_Routine_SecurityCannotDeleteOthers tests ownership checks.
func TestIntegration_Routine_SecurityCannotDeleteOthers(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	userA := srv.SeedUser(t, "routine-owner")
	userB := srv.SeedUser(t, "routine-attacker")
	tokenB := testutil.CreateTestToken(userB.ID)
	routineID := seedRoutine(t, srv, userA.ID, "Owner's Routine")

	req := httptest.NewRequest("DELETE", "/routines/"+routineID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for unauthorized deletion, got %d: %s", rr.Code, rr.Body.String())
	}
}

// =============================================================================
// RoutineExerciseHandler Tests
// =============================================================================

// TestIntegration_RoutineExercise_AddAndRemove tests adding and removing exercises from routines.
func TestIntegration_RoutineExercise_AddAndRemove(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "routine-exercise-user")
	token := testutil.CreateTestToken(user.ID)
	routineID := seedRoutine(t, srv, user.ID, "Test Routine")
	exerciseID := seedExerciseForRoutine(t, srv, user.ID, "Bench Press")

	// 1. Act - POST /routines/{id}/exercises
	payload := `{"exercise_id": "` + exerciseID.String() + `", "order_index": 1}`
	req := httptest.NewRequest("POST", "/routines/"+routineID.String()+"/exercises", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /routines/{id}/exercises: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM routine_exercises WHERE routine_id = $1",
		routineID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query routine_exercises: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 routine exercise, got %d", count)
	}

	// 3. Get routine exercise ID for deletion
	var reID uuid.UUID
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM routine_exercises WHERE routine_id = $1 LIMIT 1",
		routineID,
	).Scan(&reID)
	if err != nil {
		t.Fatalf("Failed to get routine exercise ID: %v", err)
	}

	// 4. Act - DELETE /routines/{id}/exercises/{eid}
	req = httptest.NewRequest("DELETE", "/routines/"+routineID.String()+"/exercises/"+reID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /routines/{id}/exercises/{eid}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Verify deletion
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM routine_exercises WHERE routine_id = $1",
		routineID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query routine_exercises after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 routine exercises after deletion, got %d", count)
	}
}

// =============================================================================
// RoutineSetHandler Tests
// =============================================================================

// TestIntegration_RoutineSet_AddAndRemove tests adding and removing sets from routine exercises.
func TestIntegration_RoutineSet_AddAndRemove(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "routine-set-user")
	token := testutil.CreateTestToken(user.ID)
	routineID := seedRoutine(t, srv, user.ID, "Set Routine")
	exerciseID := seedExerciseForRoutine(t, srv, user.ID, "Squat")
	reID := seedRoutineExercise(t, srv, routineID, exerciseID)

	// 1. Act - POST /routine-exercises/{id}/sets
	payload := `{"target_weight": 100.5, "target_reps": 10, "order_index": 1}`
	req := httptest.NewRequest("POST", "/routine-exercises/"+reID.String()+"/sets", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /routine-exercises/{id}/sets: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database and get set ID
	var setID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM routine_sets WHERE routine_exercise_id = $1 LIMIT 1",
		reID,
	).Scan(&setID)
	if err != nil {
		t.Fatalf("Failed to query routine_sets: %v", err)
	}

	// 3. Act - DELETE /routine-sets/{id}
	req = httptest.NewRequest("DELETE", "/routine-sets/"+setID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /routine-sets/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Verify deletion
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM routine_sets WHERE routine_exercise_id = $1",
		reID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query routine_sets after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 routine sets after deletion, got %d", count)
	}
}

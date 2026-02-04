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
// Helper: seedWorkout creates a workout for the given user
// =============================================================================

func seedWorkout(t *testing.T, srv *testutil.TestServer, userID uuid.UUID, name string) uuid.UUID {
	t.Helper()
	var workoutID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO workouts (user_id, name, started_at, ended_at, duration_seconds)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, name, time.Now().Add(-1*time.Hour), time.Now(), 3600,
	).Scan(&workoutID)
	if err != nil {
		t.Fatalf("Failed to seed workout: %v", err)
	}
	return workoutID
}

// seedExercise creates an exercise for the given user
func seedExercise(t *testing.T, srv *testutil.TestServer, userID uuid.UUID, name string) uuid.UUID {
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

// seedWorkoutExercise creates a workout exercise
func seedWorkoutExercise(t *testing.T, srv *testutil.TestServer, workoutID, exerciseID uuid.UUID) uuid.UUID {
	t.Helper()
	var weID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO workout_exercises (workout_id, exercise_id, order_index)
		 VALUES ($1, $2, 1) RETURNING id`,
		workoutID, exerciseID,
	).Scan(&weID)
	if err != nil {
		t.Fatalf("Failed to seed workout exercise: %v", err)
	}
	return weID
}

// =============================================================================
// WorkoutExerciseHandler Tests
// =============================================================================

// TestIntegration_WorkoutExercise_AddAndRemove tests adding and removing exercises from workouts.
func TestIntegration_WorkoutExercise_AddAndRemove(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "workout-exercise-user")
	token := testutil.CreateTestToken(user.ID)
	workoutID := seedWorkout(t, srv, user.ID, "Test Workout")
	exerciseID := seedExercise(t, srv, user.ID, "Bench Press")

	// 1. Act - POST /workouts/{id}/exercises
	payload := `{"exercise_id": "` + exerciseID.String() + `", "order_index": 1}`
	req := httptest.NewRequest("POST", "/workouts/"+workoutID.String()+"/exercises", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /workouts/{id}/exercises: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM workout_exercises WHERE workout_id = $1",
		workoutID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query workout_exercises: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 workout exercise, got %d", count)
	}

	// 3. Get workout exercise ID for deletion
	var weID uuid.UUID
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM workout_exercises WHERE workout_id = $1 LIMIT 1",
		workoutID,
	).Scan(&weID)
	if err != nil {
		t.Fatalf("Failed to get workout exercise ID: %v", err)
	}

	// 4. Act - DELETE /workouts/{id}/exercises/{eid}
	req = httptest.NewRequest("DELETE", "/workouts/"+workoutID.String()+"/exercises/"+weID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /workouts/{id}/exercises/{eid}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Verify deletion
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM workout_exercises WHERE workout_id = $1",
		workoutID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query workout_exercises after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 workout exercises after deletion, got %d", count)
	}
}

// TestIntegration_WorkoutExercise_SecurityCannotDeleteOthers tests ownership checks.
func TestIntegration_WorkoutExercise_SecurityCannotDeleteOthers(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	userA := srv.SeedUser(t, "owner-user")
	userB := srv.SeedUser(t, "attacker-user")
	tokenB := testutil.CreateTestToken(userB.ID)

	workoutID := seedWorkout(t, srv, userA.ID, "Owner's Workout")
	exerciseID := seedExercise(t, srv, userA.ID, "Owner's Exercise")
	weID := seedWorkoutExercise(t, srv, workoutID, exerciseID)

	// User B tries to delete User A's workout exercise
	req := httptest.NewRequest("DELETE", "/workouts/"+workoutID.String()+"/exercises/"+weID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for unauthorized deletion, got %d: %s", rr.Code, rr.Body.String())
	}
}

// =============================================================================
// WorkoutSetHandler Tests
// =============================================================================

// TestIntegration_WorkoutSet_AddUpdateRemove tests the full workout set lifecycle.
func TestIntegration_WorkoutSet_AddUpdateRemove(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "workout-set-user")
	token := testutil.CreateTestToken(user.ID)
	workoutID := seedWorkout(t, srv, user.ID, "Set Test Workout")
	exerciseID := seedExercise(t, srv, user.ID, "Squat")
	weID := seedWorkoutExercise(t, srv, workoutID, exerciseID)

	// 1. Act - POST /workout-exercises/{id}/sets
	payload := `{"weight": 100.5, "reps": 10, "order_index": 1}`
	req := httptest.NewRequest("POST", "/workout-exercises/"+weID.String()+"/sets", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /workout-exercises/{id}/sets: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var setID uuid.UUID
	var weight float64
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT id, weight FROM workout_sets WHERE workout_exercise_id = $1",
		weID,
	).Scan(&setID, &weight)
	if err != nil {
		t.Fatalf("Failed to query workout_sets: %v", err)
	}
	if weight != 100.5 {
		t.Errorf("Expected weight 100.5, got %f", weight)
	}

	// 3. Act - PUT /workout-sets/{id}
	updatePayload := `{"weight": 110.0, "reps": 8}`
	req = httptest.NewRequest("PUT", "/workout-sets/"+setID.String(), strings.NewReader(updatePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("PUT /workout-sets/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Verify update
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT weight FROM workout_sets WHERE id = $1",
		setID,
	).Scan(&weight)
	if err != nil {
		t.Fatalf("Failed to query updated workout_set: %v", err)
	}
	if weight != 110.0 {
		t.Errorf("Expected updated weight 110.0, got %f", weight)
	}

	// 5. Act - DELETE /workout-sets/{id}
	req = httptest.NewRequest("DELETE", "/workout-sets/"+setID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /workout-sets/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify deletion
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM workout_sets WHERE workout_exercise_id = $1",
		weID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query workout_sets after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 workout sets after deletion, got %d", count)
	}
}

// =============================================================================
// WorkoutImageHandler Tests
// =============================================================================

// TestIntegration_WorkoutImage_AddListRemove tests the workout image lifecycle.
func TestIntegration_WorkoutImage_AddListRemove(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "workout-image-user")
	token := testutil.CreateTestToken(user.ID)
	workoutID := seedWorkout(t, srv, user.ID, "Image Test Workout")

	// 1. Act - POST /workouts/{id}/images
	payload := `{"storage_path": "images/workout/test-image.jpg", "display_order": 1}`
	req := httptest.NewRequest("POST", "/workouts/"+workoutID.String()+"/images", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /workouts/{id}/images: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var imageID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM workout_images WHERE workout_id = $1 LIMIT 1",
		workoutID,
	).Scan(&imageID)
	if err != nil {
		t.Fatalf("Failed to query workout_images: %v", err)
	}

	// 3. Act - GET /workouts/{id}/images
	req = httptest.NewRequest("GET", "/workouts/"+workoutID.String()+"/images", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /workouts/{id}/images: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "test-image.jpg") {
		t.Errorf("Expected response to contain image path, got: %s", rr.Body.String())
	}

	// 4. Act - DELETE /workouts/{id}/images/{imageId}
	req = httptest.NewRequest("DELETE", "/workouts/"+workoutID.String()+"/images/"+imageID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /workouts/{id}/images/{imageId}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Verify deletion
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM workout_images WHERE workout_id = $1",
		workoutID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query workout_images after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 workout images after deletion, got %d", count)
	}
}

// =============================================================================
// WorkoutLikeHandler Tests
// =============================================================================

// TestIntegration_WorkoutLike_LikeListUnlike tests the workout like lifecycle.
func TestIntegration_WorkoutLike_LikeListUnlike(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	userA := srv.SeedUser(t, "workout-owner")
	userB := srv.SeedUser(t, "liker-user")
	tokenB := testutil.CreateTestToken(userB.ID)
	workoutID := seedWorkout(t, srv, userA.ID, "Likeable Workout")

	// 1. Act - POST /workouts/{id}/likes
	req := httptest.NewRequest("POST", "/workouts/"+workoutID.String()+"/likes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /workouts/{id}/likes: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM workout_likes WHERE workout_id = $1 AND user_id = $2",
		workoutID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query workout_likes: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 workout like, got %d", count)
	}

	// 3. Act - GET /workouts/{id}/likes
	req = httptest.NewRequest("GET", "/workouts/"+workoutID.String()+"/likes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /workouts/{id}/likes: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Act - DELETE /workouts/{id}/likes (unlike)
	req = httptest.NewRequest("DELETE", "/workouts/"+workoutID.String()+"/likes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /workouts/{id}/likes: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Verify unlike
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM workout_likes WHERE workout_id = $1 AND user_id = $2",
		workoutID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query workout_likes after unlike: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 workout likes after unlike, got %d", count)
	}
}

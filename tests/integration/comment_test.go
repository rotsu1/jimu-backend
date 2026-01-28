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
// Helper: seedWorkoutForComment creates a workout for comment tests
// =============================================================================

func seedWorkoutForComment(t *testing.T, srv *testutil.TestServer, userID uuid.UUID) uuid.UUID {
	t.Helper()
	var workoutID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO workouts (user_id, name, started_at, ended_at, duration_seconds)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		userID, "Comment Test Workout", time.Now().Add(-1*time.Hour), time.Now(), 3600,
	).Scan(&workoutID)
	if err != nil {
		t.Fatalf("Failed to seed workout: %v", err)
	}
	return workoutID
}

// =============================================================================
// CommentHandler Tests
// =============================================================================

// TestIntegration_Comment_Lifecycle tests creating, getting, listing, and deleting comments.
func TestIntegration_Comment_Lifecycle(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "comment-user")
	token := testutil.CreateTestToken(user.ID)
	workoutID := seedWorkoutForComment(t, srv, user.ID)

	// 1. Act - POST /comments
	payload := `{"workout_id": "` + workoutID.String() + `", "content": "Great workout!"}`
	req := httptest.NewRequest("POST", "/comments", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /comments: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database and get comment ID
	var commentID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT id FROM comments WHERE workout_id = $1 AND user_id = $2",
		workoutID, user.ID,
	).Scan(&commentID)
	if err != nil {
		t.Fatalf("Failed to query comment: %v", err)
	}

	// 3. Act - GET /comments?workout_id={id}
	req = httptest.NewRequest("GET", "/comments?workout_id="+workoutID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /comments: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "Great workout!") {
		t.Errorf("Expected response to contain comment content, got: %s", rr.Body.String())
	}

	// 4. Act - GET /comments/{id}
	req = httptest.NewRequest("GET", "/comments/"+commentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /comments/{id}: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Act - DELETE /comments/{id}
	req = httptest.NewRequest("DELETE", "/comments/"+commentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /comments/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify deletion
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM comments WHERE id = $1",
		commentID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comments after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 comments after deletion, got %d", count)
	}
}

// TestIntegration_Comment_Reply tests creating a reply to a comment.
func TestIntegration_Comment_Reply(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "reply-user")
	token := testutil.CreateTestToken(user.ID)
	workoutID := seedWorkoutForComment(t, srv, user.ID)

	// Create parent comment
	var parentID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO comments (user_id, workout_id, content)
		 VALUES ($1, $2, $3) RETURNING id`,
		user.ID, workoutID, "Parent comment",
	).Scan(&parentID)
	if err != nil {
		t.Fatalf("Failed to seed parent comment: %v", err)
	}

	// Act - Create reply
	payload := `{"workout_id": "` + workoutID.String() + `", "parent_id": "` + parentID.String() + `", "content": "Reply to parent"}`
	req := httptest.NewRequest("POST", "/comments", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /comments (reply): expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify database
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM comments WHERE parent_id = $1",
		parentID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query replies: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 reply, got %d", count)
	}
}

// TestIntegration_Comment_SecurityCannotDeleteOthers tests ownership checks.
func TestIntegration_Comment_SecurityCannotDeleteOthers(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	userA := srv.SeedUser(t, "comment-owner")
	userB := srv.SeedUser(t, "comment-attacker")
	tokenB := testutil.CreateTestToken(userB.ID)
	workoutID := seedWorkoutForComment(t, srv, userA.ID)

	// Create comment by user A
	var commentID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO comments (user_id, workout_id, content)
		 VALUES ($1, $2, $3) RETURNING id`,
		userA.ID, workoutID, "User A's comment",
	).Scan(&commentID)
	if err != nil {
		t.Fatalf("Failed to seed comment: %v", err)
	}

	// User B tries to delete User A's comment
	req := httptest.NewRequest("DELETE", "/comments/"+commentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for unauthorized deletion, got %d: %s", rr.Code, rr.Body.String())
	}
}

// =============================================================================
// CommentLikeHandler Tests
// =============================================================================

// TestIntegration_CommentLike_LikeListUnlike tests the comment like lifecycle.
func TestIntegration_CommentLike_LikeListUnlike(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	userA := srv.SeedUser(t, "comment-like-owner")
	userB := srv.SeedUser(t, "comment-liker")
	tokenB := testutil.CreateTestToken(userB.ID)
	workoutID := seedWorkoutForComment(t, srv, userA.ID)

	// Create comment by user A
	var commentID uuid.UUID
	err := srv.DB.QueryRow(
		context.Background(),
		`INSERT INTO comments (user_id, workout_id, content)
		 VALUES ($1, $2, $3) RETURNING id`,
		userA.ID, workoutID, "Likeable comment",
	).Scan(&commentID)
	if err != nil {
		t.Fatalf("Failed to seed comment: %v", err)
	}

	// 1. Act - POST /comments/{id}/likes
	req := httptest.NewRequest("POST", "/comments/"+commentID.String()+"/likes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /comments/{id}/likes: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM comment_likes WHERE comment_id = $1 AND user_id = $2",
		commentID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comment_likes: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 comment like, got %d", count)
	}

	// 3. Act - GET /comments/{id}/likes (Note: uses query param based on handler)
	req = httptest.NewRequest("GET", "/comments/"+commentID.String()+"/likes?id="+commentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /comments/{id}/likes: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Act - DELETE /comments/{id}/likes (unlike)
	req = httptest.NewRequest("DELETE", "/comments/"+commentID.String()+"/likes", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /comments/{id}/likes: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Verify unlike
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM comment_likes WHERE comment_id = $1 AND user_id = $2",
		commentID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comment_likes after unlike: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 comment likes after unlike, got %d", count)
	}
}

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rotsu1/jimu-backend/internal/testutil"
)

// =============================================================================
// FollowHandler Tests
// =============================================================================

// TestIntegration_Follow_Lifecycle tests following, listing, and unfollowing users.
func TestIntegration_Follow_Lifecycle(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	// Create two users
	userA := srv.SeedUser(t, "follow-user-a")
	userB := srv.SeedUser(t, "follow-user-b")
	tokenA := testutil.CreateTestToken(userA.ID)

	// 1. Act - POST /users/{id}/follow (A follows B)
	req := httptest.NewRequest("POST", "/users/"+userB.ID.String()+"/follow", nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /users/{id}/follow: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND following_id = $2",
		userA.ID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query follows: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 follow record, got %d", count)
	}

	// 3. Act - GET /users/{id}/followers (B's followers)
	tokenB := testutil.CreateTestToken(userB.ID)
	req = httptest.NewRequest("GET", "/users/"+userB.ID.String()+"/followers", nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /users/{id}/followers: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Act - GET /users/{id}/following (A's following)
	req = httptest.NewRequest("GET", "/users/"+userA.ID.String()+"/following", nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /users/{id}/following: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Act - DELETE /users/{id}/follow (A unfollows B)
	req = httptest.NewRequest("DELETE", "/users/"+userB.ID.String()+"/follow", nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /users/{id}/follow: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 6. Verify deletion
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND following_id = $2",
		userA.ID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query follows after deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 follow records after unfollow, got %d", count)
	}
}

// TestIntegration_Follow_CannotFollowSelf tests that users cannot follow themselves.
func TestIntegration_Follow_CannotFollowSelf(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "self-follow-user")
	token := testutil.CreateTestToken(user.ID)

	req := httptest.NewRequest("POST", "/users/"+user.ID.String()+"/follow", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for self-follow, got %d: %s", rr.Code, rr.Body.String())
	}
}

// =============================================================================
// BlockedUserHandler Tests
// =============================================================================

// TestIntegration_BlockedUser_Lifecycle tests blocking, listing, and unblocking users.
func TestIntegration_BlockedUser_Lifecycle(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	userA := srv.SeedUser(t, "blocker-user")
	userB := srv.SeedUser(t, "blocked-user")
	tokenA := testutil.CreateTestToken(userA.ID)

	// 1. Act - POST /blocked-users (A blocks B)
	payload := `{"blocked_id": "` + userB.ID.String() + `"}`
	req := httptest.NewRequest("POST", "/blocked-users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("POST /blocked-users: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM blocked_users WHERE blocker_id = $1 AND blocked_id = $2",
		userA.ID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query blocked_users: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 blocked user record, got %d", count)
	}

	// 3. Act - GET /blocked-users
	req = httptest.NewRequest("GET", "/blocked-users", nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /blocked-users: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. Act - DELETE /blocked-users/{id} (A unblocks B)
	req = httptest.NewRequest("DELETE", "/blocked-users/"+userB.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("DELETE /blocked-users/{id}: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// 5. Verify deletion
	err = srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM blocked_users WHERE blocker_id = $1 AND blocked_id = $2",
		userA.ID, userB.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query blocked_users after unblock: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 blocked user records after unblock, got %d", count)
	}
}

// =============================================================================
// SubscriptionHandler Tests
// =============================================================================

// TestIntegration_Subscription_UpsertAndGet tests subscription creation and retrieval.
func TestIntegration_Subscription_UpsertAndGet(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "subscription-user")
	token := testutil.CreateTestToken(user.ID)

	expiresAt := time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)

	// 1. Act - POST /subscriptions
	payload := `{
		"original_transaction_id": "txn_12345",
		"product_id": "premium_monthly",
		"status": "active",
		"expires_at": "` + expiresAt + `",
		"environment": "sandbox"
	}`
	req := httptest.NewRequest("POST", "/subscriptions", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("POST /subscriptions: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// 2. Verify database
	var count int
	err := srv.DB.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM subscriptions WHERE user_id = $1",
		user.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query subscriptions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 subscription, got %d", count)
	}

	// 3. Act - GET /subscriptions
	req = httptest.NewRequest("GET", "/subscriptions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /subscriptions: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "txn_12345") {
		t.Errorf("Expected response to contain transaction ID, got: %s", rr.Body.String())
	}
}

// TestIntegration_Subscription_GetNotFound tests getting subscription when none exists.
func TestIntegration_Subscription_GetNotFound(t *testing.T) {
	srv := testutil.NewTestServer(t)
	defer srv.DB.Close()

	user := srv.SeedUser(t, "no-subscription-user")
	token := testutil.CreateTestToken(user.ID)

	req := httptest.NewRequest("GET", "/subscriptions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	srv.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for no subscription, got %d: %s", rr.Code, rr.Body.String())
	}
}

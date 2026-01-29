package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rotsu1/jimu-backend/internal/handlers"
)

// explicit mock for router tests
type mockHealthRepo struct{}

func (m *mockHealthRepo) Ping(ctx context.Context) error { return nil }

// TestJimuRouter_Routing covers all routes defined in JimuRouter.
// For private routes: verifies 401 Unauthorized when no Bearer token is provided.
// For all routes: verifies 404 Not Found when wrong HTTP method is used.
func TestJimuRouter_Routing(t *testing.T) {
	// Sample UUID for path parameters
	const testUUID = "00000000-0000-0000-0000-000000000001"

	healthHandler := handlers.NewHealthHandler(&mockHealthRepo{})

	// Initialize router with nil handlers (we're testing routing, not handler logic)
	jr := &JimuRouter{
		AuthHandler:                 &handlers.AuthHandler{},
		UserSettingsHandler:         &handlers.UserSettingsHandler{},
		UserDeviceHandler:           &handlers.UserDeviceHandler{},
		SubscriptionHandler:         &handlers.SubscriptionHandler{},
		FollowHandler:               &handlers.FollowHandler{},
		BlockedUserHandler:          &handlers.BlockedUserHandler{},
		WorkoutHandler:              &handlers.WorkoutHandler{},
		WorkoutExerciseHandler:      &handlers.WorkoutExerciseHandler{},
		WorkoutSetHandler:           &handlers.WorkoutSetHandler{},
		WorkoutImageHandler:         &handlers.WorkoutImageHandler{},
		WorkoutLikeHandler:          &handlers.WorkoutLikeHandler{},
		ExerciseHandler:             &handlers.ExerciseHandler{},
		MuscleHandler:               &handlers.MuscleHandler{},
		ExerciseTargetMuscleHandler: &handlers.ExerciseTargetMuscleHandler{},
		CommentHandler:              &handlers.CommentHandler{},
		CommentLikeHandler:          &handlers.CommentLikeHandler{},
		RoutineHandler:              &handlers.RoutineHandler{},
		RoutineExerciseHandler:      &handlers.RoutineExerciseHandler{},
		RoutineSetHandler:           &handlers.RoutineSetHandler{},
		HealthHandler:               healthHandler,
		JWTSecret:                   "test-secret",
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		// =====================================================================
		// PUBLIC ROUTES
		// =====================================================================
		// Auth Login (Public)
		{"Auth Login - POST", "POST", "/auth/login", http.StatusBadRequest}, // No body, but route matched
		{"Auth Login - Wrong Method GET", "GET", "/auth/login", http.StatusNotFound},
		{"Auth Login - Wrong Method DELETE", "DELETE", "/auth/login", http.StatusNotFound},

		// Auth Refresh (Public)
		{"Auth Refresh - POST", "POST", "/auth/refresh", http.StatusBadRequest}, // No body, but route matched
		{"Auth Refresh - Wrong Method GET", "GET", "/auth/refresh", http.StatusNotFound},
		{"Auth Refresh - Wrong Method PUT", "PUT", "/auth/refresh", http.StatusNotFound},

		// Health (Public)
		{"Health Check - GET", "GET", "/health", http.StatusOK},
		{"Health Check - Wrong Method POST", "POST", "/health", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - AUTH DOMAIN
		// =====================================================================
		// Logout
		{"Logout - No Token", "POST", "/logout", http.StatusUnauthorized},
		{"Logout - Wrong Method GET", "GET", "/logout", http.StatusNotFound},

		// Auth Profile (Self)
		{"Get My Profile - No Token", "GET", "/auth/profile", http.StatusUnauthorized},
		{"Update My Profile - No Token", "PUT", "/auth/profile", http.StatusUnauthorized},
		{"Delete My Profile - No Token", "DELETE", "/auth/profile", http.StatusUnauthorized},
		{"Auth Profile - Wrong Method POST", "POST", "/auth/profile", http.StatusNotFound},

		// Auth Profile (Other User)
		{"Get Other Profile - No Token", "GET", "/auth/profile/" + testUUID, http.StatusUnauthorized},
		{"Get Other Profile - Wrong Method POST", "POST", "/auth/profile/" + testUUID, http.StatusNotFound},

		// Auth Identities
		{"Get My Identities - No Token", "GET", "/auth/identities", http.StatusUnauthorized},
		{"Unlink Identity - No Token", "DELETE", "/auth/identities/google", http.StatusUnauthorized},
		{"Auth Identities - Wrong Method POST", "POST", "/auth/identities", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - USER SETTINGS DOMAIN
		// =====================================================================
		{"Get My Settings - No Token", "GET", "/user-settings", http.StatusUnauthorized},
		{"Update My Settings - No Token", "PUT", "/user-settings", http.StatusUnauthorized},
		{"User Settings - Wrong Method POST", "POST", "/user-settings", http.StatusNotFound},
		{"User Settings - Wrong Method DELETE", "DELETE", "/user-settings", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - USER DEVICES DOMAIN
		// =====================================================================
		{"Register Device - No Token", "POST", "/user-devices", http.StatusUnauthorized},
		{"List Devices - No Token", "GET", "/user-devices", http.StatusUnauthorized},
		{"Delete Device - No Token", "DELETE", "/user-devices", http.StatusUnauthorized},
		{"User Devices - Wrong Method PUT", "PUT", "/user-devices", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - SUBSCRIPTIONS DOMAIN
		// =====================================================================
		{"Upsert Subscription - No Token", "POST", "/subscriptions", http.StatusUnauthorized},
		{"Get My Subscription - No Token", "GET", "/subscriptions", http.StatusUnauthorized},
		{"Subscriptions - Wrong Method DELETE", "DELETE", "/subscriptions", http.StatusNotFound},
		{"Subscriptions - Wrong Method PUT", "PUT", "/subscriptions", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - FOLLOWS DOMAIN (Social)
		// =====================================================================
		// Follow/Unfollow User
		{"Follow User - No Token", "POST", "/users/" + testUUID + "/follow", http.StatusUnauthorized},
		{"Unfollow User - No Token", "DELETE", "/users/" + testUUID + "/follow", http.StatusUnauthorized},
		{"Follow - Wrong Method GET", "GET", "/users/" + testUUID + "/follow", http.StatusNotFound},

		// Followers/Following Lists
		{"Get Followers - No Token", "GET", "/users/" + testUUID + "/followers", http.StatusUnauthorized},
		{"Get Following - No Token", "GET", "/users/" + testUUID + "/following", http.StatusUnauthorized},
		{"Followers - Wrong Method POST", "POST", "/users/" + testUUID + "/followers", http.StatusNotFound},
		{"Following - Wrong Method DELETE", "DELETE", "/users/" + testUUID + "/following", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - BLOCKED USERS DOMAIN (Social)
		// =====================================================================
		{"Block User - No Token", "POST", "/blocked-users", http.StatusUnauthorized},
		{"Get Blocked Users - No Token", "GET", "/blocked-users", http.StatusUnauthorized},
		{"Blocked Users - Wrong Method DELETE on collection", "DELETE", "/blocked-users", http.StatusNotFound},

		// Unblock User (with ID)
		{"Unblock User - No Token", "DELETE", "/blocked-users/" + testUUID, http.StatusUnauthorized},
		{"Unblock User - Wrong Method GET", "GET", "/blocked-users/" + testUUID, http.StatusNotFound},
		{"Unblock User - Wrong Method POST", "POST", "/blocked-users/" + testUUID, http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - WORKOUTS DOMAIN
		// =====================================================================
		// Collection routes
		{"List Workouts - No Token", "GET", "/workouts", http.StatusUnauthorized},
		{"Create Workout - No Token", "POST", "/workouts", http.StatusUnauthorized},
		{"Workouts Collection - Wrong Method DELETE", "DELETE", "/workouts", http.StatusNotFound},
		{"Workouts Collection - Wrong Method PUT", "PUT", "/workouts", http.StatusNotFound},

		// Single workout routes
		{"Get Workout - No Token", "GET", "/workouts/" + testUUID, http.StatusUnauthorized},
		{"Update Workout - No Token", "PUT", "/workouts/" + testUUID, http.StatusUnauthorized},
		{"Delete Workout - No Token", "DELETE", "/workouts/" + testUUID, http.StatusUnauthorized},
		{"Workout Detail - Wrong Method POST", "POST", "/workouts/" + testUUID, http.StatusNotFound},

		// Workout Likes sub-resource
		{"Like Workout - No Token", "POST", "/workouts/" + testUUID + "/likes", http.StatusUnauthorized},
		{"Unlike Workout - No Token", "DELETE", "/workouts/" + testUUID + "/likes", http.StatusUnauthorized},
		{"List Workout Likes - No Token", "GET", "/workouts/" + testUUID + "/likes", http.StatusUnauthorized},
		{"Workout Likes - Wrong Method PUT", "PUT", "/workouts/" + testUUID + "/likes", http.StatusNotFound},

		// Workout Images sub-resource
		{"Add Workout Image - No Token", "POST", "/workouts/" + testUUID + "/images", http.StatusUnauthorized},
		{"List Workout Images - No Token", "GET", "/workouts/" + testUUID + "/images", http.StatusUnauthorized},
		{"Workout Images - Wrong Method DELETE on collection", "DELETE", "/workouts/" + testUUID + "/images", http.StatusNotFound},
		{"Remove Workout Image - No Token", "DELETE", "/workouts/" + testUUID + "/images/" + testUUID, http.StatusUnauthorized},
		{"Remove Workout Image - Wrong Method GET", "GET", "/workouts/" + testUUID + "/images/" + testUUID, http.StatusNotFound},

		// Workout Exercises sub-resource
		{"Add Workout Exercise - No Token", "POST", "/workouts/" + testUUID + "/exercises", http.StatusUnauthorized},
		{"Workout Exercises - Wrong Method GET", "GET", "/workouts/" + testUUID + "/exercises", http.StatusNotFound},
		{"Update Workout Exercise - No Token", "PUT", "/workouts/" + testUUID + "/exercises/" + testUUID, http.StatusUnauthorized},
		{"Remove Workout Exercise - No Token", "DELETE", "/workouts/" + testUUID + "/exercises/" + testUUID, http.StatusUnauthorized},
		{"Workout Exercise Detail - Wrong Method GET", "GET", "/workouts/" + testUUID + "/exercises/" + testUUID, http.StatusNotFound},
		{"Workout Exercise Detail - Wrong Method POST", "POST", "/workouts/" + testUUID + "/exercises/" + testUUID, http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - WORKOUT SETS (Standalone)
		// =====================================================================
		{"Update Workout Set - No Token", "PUT", "/workout-sets/" + testUUID, http.StatusUnauthorized},
		{"Remove Workout Set - No Token", "DELETE", "/workout-sets/" + testUUID, http.StatusUnauthorized},
		{"Workout Sets - Wrong Method GET", "GET", "/workout-sets/" + testUUID, http.StatusNotFound},
		{"Workout Sets - Wrong Method POST", "POST", "/workout-sets/" + testUUID, http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - WORKOUT EXERCISES (Sets sub-resource)
		// =====================================================================
		{"Add Set to Workout Exercise - No Token", "POST", "/workout-exercises/" + testUUID + "/sets", http.StatusUnauthorized},
		{"Workout Exercise Sets - Wrong Method GET", "GET", "/workout-exercises/" + testUUID + "/sets", http.StatusNotFound},
		{"Workout Exercise Sets - Wrong Method DELETE", "DELETE", "/workout-exercises/" + testUUID + "/sets", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - EXERCISES DOMAIN
		// =====================================================================
		// Collection routes
		{"List Exercises - No Token", "GET", "/exercises", http.StatusUnauthorized},
		{"Create Exercise - No Token", "POST", "/exercises", http.StatusUnauthorized},
		{"Exercises Collection - Wrong Method DELETE", "DELETE", "/exercises", http.StatusNotFound},

		// Single exercise routes
		{"Get Exercise - No Token", "GET", "/exercises/" + testUUID, http.StatusUnauthorized},
		{"Update Exercise - No Token", "PUT", "/exercises/" + testUUID, http.StatusUnauthorized},
		{"Delete Exercise - No Token", "DELETE", "/exercises/" + testUUID, http.StatusUnauthorized},
		{"Exercise Detail - Wrong Method POST", "POST", "/exercises/" + testUUID, http.StatusNotFound},

		// Exercise Target Muscles sub-resource
		{"Add Target Muscle - No Token", "POST", "/exercises/" + testUUID + "/muscles", http.StatusUnauthorized},
		{"Exercise Muscles - Wrong Method GET", "GET", "/exercises/" + testUUID + "/muscles", http.StatusNotFound},
		{"Remove Target Muscle - No Token", "DELETE", "/exercises/" + testUUID + "/muscles/" + testUUID, http.StatusUnauthorized},
		{"Exercise Muscle Detail - Wrong Method GET", "GET", "/exercises/" + testUUID + "/muscles/" + testUUID, http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - MUSCLES DOMAIN
		// =====================================================================
		// Collection routes
		{"List Muscles - No Token", "GET", "/muscles", http.StatusUnauthorized},
		{"Create Muscle - No Token", "POST", "/muscles", http.StatusUnauthorized},
		{"Muscles Collection - Wrong Method DELETE", "DELETE", "/muscles", http.StatusNotFound},
		{"Muscles Collection - Wrong Method PUT", "PUT", "/muscles", http.StatusNotFound},

		// Single muscle routes
		{"Get Muscle - No Token", "GET", "/muscles/" + testUUID, http.StatusUnauthorized},
		{"Delete Muscle - No Token", "DELETE", "/muscles/" + testUUID, http.StatusUnauthorized},
		{"Muscle Detail - Wrong Method POST", "POST", "/muscles/" + testUUID, http.StatusNotFound},
		{"Muscle Detail - Wrong Method PUT", "PUT", "/muscles/" + testUUID, http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - COMMENTS DOMAIN
		// =====================================================================
		// Collection routes
		{"List Comments - No Token", "GET", "/comments", http.StatusUnauthorized},
		{"Create Comment - No Token", "POST", "/comments", http.StatusUnauthorized},
		{"Comments Collection - Wrong Method DELETE", "DELETE", "/comments", http.StatusNotFound},
		{"Comments Collection - Wrong Method PUT", "PUT", "/comments", http.StatusNotFound},

		// Single comment routes
		{"Get Comment - No Token", "GET", "/comments/" + testUUID, http.StatusUnauthorized},
		{"Delete Comment - No Token", "DELETE", "/comments/" + testUUID, http.StatusUnauthorized},
		{"Comment Detail - Wrong Method POST", "POST", "/comments/" + testUUID, http.StatusNotFound},
		{"Comment Detail - Wrong Method PUT", "PUT", "/comments/" + testUUID, http.StatusNotFound},

		// Comment Likes sub-resource
		{"Like Comment - No Token", "POST", "/comments/" + testUUID + "/likes", http.StatusUnauthorized},
		{"Unlike Comment - No Token", "DELETE", "/comments/" + testUUID + "/likes", http.StatusUnauthorized},
		{"List Comment Likes - No Token", "GET", "/comments/" + testUUID + "/likes", http.StatusUnauthorized},
		{"Comment Likes - Wrong Method PUT", "PUT", "/comments/" + testUUID + "/likes", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - ROUTINES DOMAIN
		// =====================================================================
		// Collection routes
		{"List Routines - No Token", "GET", "/routines", http.StatusUnauthorized},
		{"Create Routine - No Token", "POST", "/routines", http.StatusUnauthorized},
		{"Routines Collection - Wrong Method DELETE", "DELETE", "/routines", http.StatusNotFound},
		{"Routines Collection - Wrong Method PUT", "PUT", "/routines", http.StatusNotFound},

		// Single routine routes
		{"Get Routine - No Token", "GET", "/routines/" + testUUID, http.StatusUnauthorized},
		{"Update Routine - No Token", "PUT", "/routines/" + testUUID, http.StatusUnauthorized},
		{"Delete Routine - No Token", "DELETE", "/routines/" + testUUID, http.StatusUnauthorized},
		{"Routine Detail - Wrong Method POST", "POST", "/routines/" + testUUID, http.StatusNotFound},

		// Routine Exercises sub-resource
		{"Add Routine Exercise - No Token", "POST", "/routines/" + testUUID + "/exercises", http.StatusUnauthorized},
		{"Routine Exercises - Wrong Method GET", "GET", "/routines/" + testUUID + "/exercises", http.StatusNotFound},
		{"Remove Routine Exercise - No Token", "DELETE", "/routines/" + testUUID + "/exercises/" + testUUID, http.StatusUnauthorized},
		{"Routine Exercise Detail - Wrong Method GET", "GET", "/routines/" + testUUID + "/exercises/" + testUUID, http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - ROUTINE EXERCISES (Sets sub-resource)
		// =====================================================================
		{"Add Set to Routine Exercise - No Token", "POST", "/routine-exercises/" + testUUID + "/sets", http.StatusUnauthorized},
		{"Routine Exercise Sets - Wrong Method GET", "GET", "/routine-exercises/" + testUUID + "/sets", http.StatusNotFound},
		{"Routine Exercise Sets - Wrong Method DELETE", "DELETE", "/routine-exercises/" + testUUID + "/sets", http.StatusNotFound},

		// =====================================================================
		// PRIVATE ROUTES - ROUTINE SETS (Standalone)
		// =====================================================================
		{"Remove Routine Set - No Token", "DELETE", "/routine-sets/" + testUUID, http.StatusUnauthorized},
		{"Routine Sets - Wrong Method GET", "GET", "/routine-sets/" + testUUID, http.StatusNotFound},
		{"Routine Sets - Wrong Method POST", "POST", "/routine-sets/" + testUUID, http.StatusNotFound},
		{"Routine Sets - Wrong Method PUT", "PUT", "/routine-sets/" + testUUID, http.StatusNotFound},

		// =====================================================================
		// NON-EXISTENT ROUTES (Should 404)
		// =====================================================================
		{"Non-existent route - GET /nonexistent", "GET", "/nonexistent", http.StatusNotFound},
		{"Non-existent route - POST /foo/bar", "POST", "/foo/bar", http.StatusNotFound},
		{"Non-existent route - GET /workouts/invalid/subresource", "GET", "/workouts/" + testUUID + "/invalid", http.StatusNotFound},
		{"Non-existent route - GET /routines/invalid/subresource", "GET", "/routines/" + testUUID + "/invalid", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			jr.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("%s %s: got status %d, want %d", tt.method, tt.path, rec.Code, tt.expectedStatus)
			}
		})
	}
}

// TestJimuRouter_PathParsingWithVariousUUIDs ensures path parsing works correctly
// with different valid UUID formats in the path.
func TestJimuRouter_PathParsingWithVariousUUIDs(t *testing.T) {
	healthHandler := handlers.NewHealthHandler(&mockHealthRepo{})
	jr := &JimuRouter{
		AuthHandler:                 &handlers.AuthHandler{},
		UserSettingsHandler:         &handlers.UserSettingsHandler{},
		UserDeviceHandler:           &handlers.UserDeviceHandler{},
		SubscriptionHandler:         &handlers.SubscriptionHandler{},
		FollowHandler:               &handlers.FollowHandler{},
		BlockedUserHandler:          &handlers.BlockedUserHandler{},
		WorkoutHandler:              &handlers.WorkoutHandler{},
		WorkoutExerciseHandler:      &handlers.WorkoutExerciseHandler{},
		WorkoutSetHandler:           &handlers.WorkoutSetHandler{},
		WorkoutImageHandler:         &handlers.WorkoutImageHandler{},
		WorkoutLikeHandler:          &handlers.WorkoutLikeHandler{},
		ExerciseHandler:             &handlers.ExerciseHandler{},
		MuscleHandler:               &handlers.MuscleHandler{},
		ExerciseTargetMuscleHandler: &handlers.ExerciseTargetMuscleHandler{},
		CommentHandler:              &handlers.CommentHandler{},
		CommentLikeHandler:          &handlers.CommentLikeHandler{},
		RoutineHandler:              &handlers.RoutineHandler{},
		RoutineExerciseHandler:      &handlers.RoutineExerciseHandler{},
		RoutineSetHandler:           &handlers.RoutineSetHandler{},
		HealthHandler:               healthHandler,
		JWTSecret:                   "test-secret",
	}

	uuids := []string{
		"00000000-0000-0000-0000-000000000001",
		"550e8400-e29b-41d4-a716-446655440000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
	}

	for _, uuid := range uuids {
		// Each UUID should route to the correct handler (returns 401 because no token)
		paths := []string{
			"/workouts/" + uuid,
			"/workouts/" + uuid + "/likes",
			"/exercises/" + uuid,
			"/routines/" + uuid,
			"/comments/" + uuid,
			"/blocked-users/" + uuid,
		}

		for _, path := range paths {
			t.Run("GET "+path, func(t *testing.T) {
				req := httptest.NewRequest("GET", path, nil)
				rec := httptest.NewRecorder()
				jr.ServeHTTP(rec, req)

				// All private routes should return 401 without token
				// Exception: blocked-users/{id} only has DELETE, not GET, so expect 404
				if path == "/blocked-users/"+uuid {
					if rec.Code != http.StatusNotFound {
						t.Errorf("GET %s: got status %d, want %d", path, rec.Code, http.StatusNotFound)
					}
				} else {
					if rec.Code != http.StatusUnauthorized {
						t.Errorf("GET %s: got status %d, want %d", path, rec.Code, http.StatusUnauthorized)
					}
				}
			})
		}
	}
}

// TestJimuRouter_SubResourceIsolation ensures sub-resource routes don't collide
// with parent routes.
func TestJimuRouter_SubResourceIsolation(t *testing.T) {
	healthHandler := handlers.NewHealthHandler(&mockHealthRepo{})
	jr := &JimuRouter{
		AuthHandler:                 &handlers.AuthHandler{},
		UserSettingsHandler:         &handlers.UserSettingsHandler{},
		UserDeviceHandler:           &handlers.UserDeviceHandler{},
		SubscriptionHandler:         &handlers.SubscriptionHandler{},
		FollowHandler:               &handlers.FollowHandler{},
		BlockedUserHandler:          &handlers.BlockedUserHandler{},
		WorkoutHandler:              &handlers.WorkoutHandler{},
		WorkoutExerciseHandler:      &handlers.WorkoutExerciseHandler{},
		WorkoutSetHandler:           &handlers.WorkoutSetHandler{},
		WorkoutImageHandler:         &handlers.WorkoutImageHandler{},
		WorkoutLikeHandler:          &handlers.WorkoutLikeHandler{},
		ExerciseHandler:             &handlers.ExerciseHandler{},
		MuscleHandler:               &handlers.MuscleHandler{},
		ExerciseTargetMuscleHandler: &handlers.ExerciseTargetMuscleHandler{},
		CommentHandler:              &handlers.CommentHandler{},
		CommentLikeHandler:          &handlers.CommentLikeHandler{},
		RoutineHandler:              &handlers.RoutineHandler{},
		RoutineExerciseHandler:      &handlers.RoutineExerciseHandler{},
		RoutineSetHandler:           &handlers.RoutineSetHandler{},
		HealthHandler:               healthHandler,
		JWTSecret:                   "test-secret",
	}

	const testUUID = "00000000-0000-0000-0000-000000000001"

	// Test that accessing different sub-resources on the same parent routes correctly
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		// Parent route should not match sub-resource paths
		{"Workout parent GET", "GET", "/workouts/" + testUUID, http.StatusUnauthorized},
		{"Workout likes sub-resource POST", "POST", "/workouts/" + testUUID + "/likes", http.StatusUnauthorized},
		{"Workout images sub-resource POST", "POST", "/workouts/" + testUUID + "/images", http.StatusUnauthorized},
		{"Workout exercises sub-resource POST", "POST", "/workouts/" + testUUID + "/exercises", http.StatusUnauthorized},

		// Nested resources with secondary IDs
		{"Workout image removal", "DELETE", "/workouts/" + testUUID + "/images/" + testUUID, http.StatusUnauthorized},
		{"Workout exercise removal", "DELETE", "/workouts/" + testUUID + "/exercises/" + testUUID, http.StatusUnauthorized},

		// Routine similar tests
		{"Routine parent GET", "GET", "/routines/" + testUUID, http.StatusUnauthorized},
		{"Routine exercises sub-resource POST", "POST", "/routines/" + testUUID + "/exercises", http.StatusUnauthorized},
		{"Routine exercise removal", "DELETE", "/routines/" + testUUID + "/exercises/" + testUUID, http.StatusUnauthorized},

		// Comments with likes
		{"Comment parent GET", "GET", "/comments/" + testUUID, http.StatusUnauthorized},
		{"Comment likes sub-resource POST", "POST", "/comments/" + testUUID + "/likes", http.StatusUnauthorized},

		// Exercise with muscles
		{"Exercise parent GET", "GET", "/exercises/" + testUUID, http.StatusUnauthorized},
		{"Exercise muscles sub-resource POST", "POST", "/exercises/" + testUUID + "/muscles", http.StatusUnauthorized},
		{"Exercise muscle removal", "DELETE", "/exercises/" + testUUID + "/muscles/" + testUUID, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			jr.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("%s %s: got status %d, want %d", tt.method, tt.path, rec.Code, tt.expectedStatus)
			}
		})
	}
}

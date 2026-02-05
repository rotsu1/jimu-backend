package router

import (
	"net/http"
	"strings"

	"github.com/rotsu1/jimu-backend/internal/handlers"
	"github.com/rotsu1/jimu-backend/internal/middleware"
)

type JimuRouter struct {
	AuthHandler                 *handlers.AuthHandler
	UserSettingsHandler         *handlers.UserSettingsHandler
	UserDeviceHandler           *handlers.UserDeviceHandler
	SubscriptionHandler         *handlers.SubscriptionHandler
	FollowHandler               *handlers.FollowHandler
	BlockedUserHandler          *handlers.BlockedUserHandler
	WorkoutHandler              *handlers.WorkoutHandler
	WorkoutExerciseHandler      *handlers.WorkoutExerciseHandler
	WorkoutSetHandler           *handlers.WorkoutSetHandler
	WorkoutImageHandler         *handlers.WorkoutImageHandler
	WorkoutLikeHandler          *handlers.WorkoutLikeHandler
	ExerciseHandler             *handlers.ExerciseHandler
	MuscleHandler               *handlers.MuscleHandler
	ExerciseTargetMuscleHandler *handlers.ExerciseTargetMuscleHandler
	CommentHandler              *handlers.CommentHandler
	CommentLikeHandler          *handlers.CommentLikeHandler
	RoutineHandler              *handlers.RoutineHandler
	RoutineExerciseHandler      *handlers.RoutineExerciseHandler
	RoutineSetHandler           *handlers.RoutineSetHandler
	HealthHandler               *handlers.HealthHandler
	JWTSecret                   string
}

func (jr *JimuRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	// --- 1. Public Routes ---
	switch path {
	case "/auth/login":
		if method == "POST" {
			jr.AuthHandler.GoogleLogin(w, r)
			return
		}
	case "/auth/refresh":
		if method == "POST" {
			jr.AuthHandler.RefreshToken(w, r)
			return
		}
	case "/health":
		if method == "GET" {
			jr.HealthHandler.HealthCheck(w, r)
			return
		}
	}

	// --- 2. Private Routes ---
	authMW := middleware.AuthMiddleware(jr.JWTSecret)

	// --- Auth Routes ---
	if strings.HasPrefix(path, "/logout") {
		if method == "POST" {
			authMW(http.HandlerFunc(jr.AuthHandler.Logout)).ServeHTTP(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/auth/profile") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if method == "GET" {
			// GET /auth/profile/{id} -> GetOtherProfile
			// GET /auth/profile -> GetMyProfile
			if len(parts) == 3 {
				authMW(http.HandlerFunc(jr.AuthHandler.GetOtherProfile)).ServeHTTP(w, r)
			} else {
				authMW(http.HandlerFunc(jr.AuthHandler.GetMyProfile)).ServeHTTP(w, r)
			}
			return
		}
		if method == "PUT" {
			authMW(http.HandlerFunc(jr.AuthHandler.UpdateMyProfile)).ServeHTTP(w, r)
			return
		}
		if method == "DELETE" {
			authMW(http.HandlerFunc(jr.AuthHandler.DeleteMyProfile)).ServeHTTP(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/auth/identities") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		// /auth/identities
		if len(parts) == 2 {
			if method == "GET" {
				authMW(http.HandlerFunc(jr.AuthHandler.GetMyIdentities)).ServeHTTP(w, r)
				return
			}
		}
		// /auth/identities/{provider}
		if len(parts) == 3 {
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.AuthHandler.UnlinkIdentity)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- User Settings Routes ---
	if strings.HasPrefix(path, "/user-settings") {
		if method == "GET" {
			authMW(http.HandlerFunc(jr.UserSettingsHandler.GetMySettings)).ServeHTTP(w, r)
			return
		}
		if method == "PUT" {
			authMW(http.HandlerFunc(jr.UserSettingsHandler.UpdateMySettings)).ServeHTTP(w, r)
			return
		}
	}

	// --- User Devices Routes ---
	if strings.HasPrefix(path, "/user-devices") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 1 {
			// /user-devices
			if method == "POST" {
				authMW(http.HandlerFunc(jr.UserDeviceHandler.RegisterDevice)).ServeHTTP(w, r)
				return
			}
			if method == "GET" {
				authMW(http.HandlerFunc(jr.UserDeviceHandler.ListDevices)).ServeHTTP(w, r)
				return
			}
		}
		if len(parts) == 2 {
			// /user-devices/{id}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.UserDeviceHandler.DeleteDevice)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- Subscription Routes ---
	// POST /subscriptions -> Upsert
	// GET /subscriptions -> GetMySubscription
	if strings.HasPrefix(path, "/subscriptions") {
		if method == "POST" {
			authMW(http.HandlerFunc(jr.SubscriptionHandler.UpsertSubscription)).ServeHTTP(w, r)
			return
		}
		if method == "GET" {
			authMW(http.HandlerFunc(jr.SubscriptionHandler.GetMySubscription)).ServeHTTP(w, r)
			return
		}
	}

	// --- Follow Routes ---
	// POST /users/{id}/follow -> FollowUser
	// DELETE /users/{id}/follow -> UnfollowUser
	// GET /users/{id}/followers -> GetFollowers
	// GET /users/{id}/following -> GetFollowing
	if strings.HasPrefix(path, "/users/") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		// parts: ["users", "{id}", "follow|followers|following"]
		if len(parts) >= 3 {
			switch parts[2] {
			case "follow":
				if method == "POST" {
					authMW(http.HandlerFunc(jr.FollowHandler.FollowUser)).ServeHTTP(w, r)
					return
				}
				if method == "DELETE" {
					authMW(http.HandlerFunc(jr.FollowHandler.UnfollowUser)).ServeHTTP(w, r)
					return
				}
			case "followers":
				if method == "GET" {
					authMW(http.HandlerFunc(jr.FollowHandler.GetFollowers)).ServeHTTP(w, r)
					return
				}
			case "following":
				if method == "GET" {
					authMW(http.HandlerFunc(jr.FollowHandler.GetFollowing)).ServeHTTP(w, r)
					return
				}
			}
		}
	}

	// --- Blocked Users Routes ---
	// POST /blocked-users -> BlockUser
	// GET /blocked-users -> GetBlockedUsers
	// DELETE /blocked-users/{id} -> UnblockUser
	if strings.HasPrefix(path, "/blocked-users") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 1 {
			// /blocked-users
			if method == "POST" {
				authMW(http.HandlerFunc(jr.BlockedUserHandler.BlockUser)).ServeHTTP(w, r)
				return
			}
			if method == "GET" {
				authMW(http.HandlerFunc(jr.BlockedUserHandler.GetBlockedUsers)).ServeHTTP(w, r)
				return
			}
		}
		if len(parts) == 2 {
			// /blocked-users/{id}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.BlockedUserHandler.UnblockUser)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- Workout Routes (with sub-resources) ---
	// GET /workouts -> ListWorkouts
	// POST /workouts -> CreateWorkout
	// GET /workouts/{id} -> GetWorkout
	// PUT /workouts/{id} -> UpdateWorkout
	// DELETE /workouts/{id} -> DeleteWorkout
	// POST /workouts/{id}/likes -> LikeWorkout
	// DELETE /workouts/{id}/likes -> UnlikeWorkout
	// GET /workouts/{id}/likes -> ListLikes
	// POST /workouts/{id}/images -> AddImage
	// GET /workouts/{id}/images -> ListImages
	// DELETE /workouts/{id}/images/{imageId} -> RemoveImage
	// POST /workouts/{id}/exercises -> AddExercise
	// PUT /workouts/{id}/exercises/{exerciseId} -> UpdateExercise
	// DELETE /workouts/{id}/exercises/{exerciseId} -> RemoveExercise
	// POST /workouts/{id}/sets -> AddSet (via workout-exercises)
	// DELETE /workouts/{id}/sets/{setId} -> RemoveSet
	// PUT /workouts/{id}/sets/{setId} -> UpdateSet
	if strings.HasPrefix(path, "/workouts") {
		parts := strings.Split(strings.Trim(path, "/"), "/")

		// GET /workouts or POST /workouts
		if len(parts) == 1 {
			if method == "GET" {
				authMW(http.HandlerFunc(jr.WorkoutHandler.ListWorkouts)).ServeHTTP(w, r)
				return
			}
			if method == "POST" {
				authMW(http.HandlerFunc(jr.WorkoutHandler.CreateWorkout)).ServeHTTP(w, r)
				return
			}
		}

		// GET /workouts/timeline -> GetTimelineWorkouts (query: user_id, limit, offset)
		if len(parts) == 2 && parts[1] == "timeline" {
			if method == "GET" {
				authMW(http.HandlerFunc(jr.WorkoutHandler.GetTimelineWorkouts)).ServeHTTP(w, r)
				return
			}
		}

		// /workouts/{id} (exclude "timeline" which is handled above)
		if len(parts) == 2 && parts[1] != "timeline" {
			if method == "GET" {
				authMW(http.HandlerFunc(jr.WorkoutHandler.GetWorkout)).ServeHTTP(w, r)
				return
			}
			if method == "PUT" {
				authMW(http.HandlerFunc(jr.WorkoutHandler.UpdateWorkout)).ServeHTTP(w, r)
				return
			}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.WorkoutHandler.DeleteWorkout)).ServeHTTP(w, r)
				return
			}
		}

		// /workouts/{id}/{sub-resource}
		if len(parts) >= 3 {
			subResource := parts[2]
			switch subResource {
			case "likes":
				// POST /workouts/{id}/likes -> LikeWorkout
				// DELETE /workouts/{id}/likes -> UnlikeWorkout
				// GET /workouts/{id}/likes -> ListLikes
				if len(parts) == 3 {
					if method == "POST" {
						authMW(http.HandlerFunc(jr.WorkoutLikeHandler.LikeWorkout)).ServeHTTP(w, r)
						return
					}
					if method == "DELETE" {
						authMW(http.HandlerFunc(jr.WorkoutLikeHandler.UnlikeWorkout)).ServeHTTP(w, r)
						return
					}
					if method == "GET" {
						authMW(http.HandlerFunc(jr.WorkoutLikeHandler.ListLikes)).ServeHTTP(w, r)
						return
					}
				}
			case "images":
				// POST /workouts/{id}/images -> AddImage
				// GET /workouts/{id}/images -> ListImages
				if len(parts) == 3 {
					if method == "POST" {
						authMW(http.HandlerFunc(jr.WorkoutImageHandler.AddImage)).ServeHTTP(w, r)
						return
					}
					if method == "GET" {
						authMW(http.HandlerFunc(jr.WorkoutImageHandler.ListImages)).ServeHTTP(w, r)
						return
					}
				}
				// DELETE /workouts/{id}/images/{imageId} -> RemoveImage
				if len(parts) == 4 {
					if method == "DELETE" {
						authMW(http.HandlerFunc(jr.WorkoutImageHandler.RemoveImage)).ServeHTTP(w, r)
						return
					}
				}
			case "exercises":
				// POST /workouts/{id}/exercises -> AddExercise
				if len(parts) == 3 {
					if method == "POST" {
						authMW(http.HandlerFunc(jr.WorkoutExerciseHandler.AddExercise)).ServeHTTP(w, r)
						return
					}
				}
				// PUT /workouts/{id}/exercises/{exerciseId} -> UpdateExercise
				// DELETE /workouts/{id}/exercises/{exerciseId} -> RemoveExercise
				if len(parts) == 4 {
					if method == "PUT" {
						authMW(http.HandlerFunc(jr.WorkoutExerciseHandler.UpdateExercise)).ServeHTTP(w, r)
						return
					}
					if method == "DELETE" {
						authMW(http.HandlerFunc(jr.WorkoutExerciseHandler.RemoveExercise)).ServeHTTP(w, r)
						return
					}
				}
			}
		}
	}

	// --- Workout Sets Routes (standalone) ---
	// PUT /workout-sets/{id} -> UpdateSet
	// DELETE /workout-sets/{id} -> RemoveSet
	if strings.HasPrefix(path, "/workout-sets") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 2 {
			if method == "PUT" {
				authMW(http.HandlerFunc(jr.WorkoutSetHandler.UpdateSet)).ServeHTTP(w, r)
				return
			}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.WorkoutSetHandler.RemoveSet)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- Workout Exercises Sub-routes (for sets) ---
	// POST /workout-exercises/{id}/sets -> AddSet
	if strings.HasPrefix(path, "/workout-exercises") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		// /workout-exercises/{id}/sets
		if len(parts) == 3 && parts[2] == "sets" {
			if method == "POST" {
				authMW(http.HandlerFunc(jr.WorkoutSetHandler.AddSet)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- Exercise Routes ---
	// GET /exercises -> ListExercises
	// POST /exercises -> CreateExercise
	// GET /exercises/{id} -> GetExercise
	// PUT /exercises/{id} -> UpdateExercise
	// DELETE /exercises/{id} -> DeleteExercise
	// POST /exercises/{id}/muscles -> AddTargetMuscle
	// DELETE /exercises/{id}/muscles/{muscleId} -> RemoveTargetMuscle
	if strings.HasPrefix(path, "/exercises") {
		parts := strings.Split(strings.Trim(path, "/"), "/")

		if len(parts) == 1 {
			// GET /exercises or POST /exercises
			if method == "GET" {
				authMW(http.HandlerFunc(jr.ExerciseHandler.ListExercises)).ServeHTTP(w, r)
				return
			}
			if method == "POST" {
				authMW(http.HandlerFunc(jr.ExerciseHandler.CreateExercise)).ServeHTTP(w, r)
				return
			}
		}

		if len(parts) == 2 {
			// GET/PUT/DELETE /exercises/{id}
			if method == "GET" {
				authMW(http.HandlerFunc(jr.ExerciseHandler.GetExercise)).ServeHTTP(w, r)
				return
			}
			if method == "PUT" {
				authMW(http.HandlerFunc(jr.ExerciseHandler.UpdateExercise)).ServeHTTP(w, r)
				return
			}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.ExerciseHandler.DeleteExercise)).ServeHTTP(w, r)
				return
			}
		}

		// /exercises/{id}/muscles
		if len(parts) >= 3 && parts[2] == "muscles" {
			if len(parts) == 3 {
				// POST /exercises/{id}/muscles -> AddTargetMuscle
				if method == "POST" {
					authMW(http.HandlerFunc(jr.ExerciseTargetMuscleHandler.AddTargetMuscle)).ServeHTTP(w, r)
					return
				}
			}
			if len(parts) == 4 {
				// DELETE /exercises/{id}/muscles/{muscleId} -> RemoveTargetMuscle
				if method == "DELETE" {
					authMW(http.HandlerFunc(jr.ExerciseTargetMuscleHandler.RemoveTargetMuscle)).ServeHTTP(w, r)
					return
				}
			}
		}
	}

	// --- Muscle Routes ---
	// GET /muscles -> ListMuscles
	// POST /muscles -> CreateMuscle
	// GET /muscles/{id} -> GetMuscle
	// DELETE /muscles/{id} -> DeleteMuscle
	if strings.HasPrefix(path, "/muscles") {
		parts := strings.Split(strings.Trim(path, "/"), "/")

		if len(parts) == 1 {
			if method == "GET" {
				authMW(http.HandlerFunc(jr.MuscleHandler.ListMuscles)).ServeHTTP(w, r)
				return
			}
			if method == "POST" {
				authMW(http.HandlerFunc(jr.MuscleHandler.CreateMuscle)).ServeHTTP(w, r)
				return
			}
		}

		if len(parts) == 2 {
			if method == "GET" {
				authMW(http.HandlerFunc(jr.MuscleHandler.GetMuscle)).ServeHTTP(w, r)
				return
			}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.MuscleHandler.DeleteMuscle)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- Comment Routes ---
	// GET /comments -> ListComments (with query params: workout_id, parent_id)
	// POST /comments -> CreateComment
	// GET /comments/{id} -> GetComment
	// DELETE /comments/{id} -> DeleteComment
	// POST /comments/{id}/likes -> LikeComment
	// DELETE /comments/{id}/likes -> UnlikeComment
	// GET /comments/{id}/likes -> ListLikes (for comment)
	if strings.HasPrefix(path, "/comments") {
		parts := strings.Split(strings.Trim(path, "/"), "/")

		if len(parts) == 1 {
			// GET /comments or POST /comments
			if method == "GET" {
				authMW(http.HandlerFunc(jr.CommentHandler.ListComments)).ServeHTTP(w, r)
				return
			}
			if method == "POST" {
				authMW(http.HandlerFunc(jr.CommentHandler.CreateComment)).ServeHTTP(w, r)
				return
			}
		}

		if len(parts) == 2 {
			// GET/DELETE /comments/{id}
			if method == "GET" {
				authMW(http.HandlerFunc(jr.CommentHandler.GetComment)).ServeHTTP(w, r)
				return
			}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.CommentHandler.DeleteComment)).ServeHTTP(w, r)
				return
			}
		}

		// /comments/{id}/likes
		if len(parts) == 3 && parts[2] == "likes" {
			if method == "POST" {
				authMW(http.HandlerFunc(jr.CommentLikeHandler.LikeComment)).ServeHTTP(w, r)
				return
			}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.CommentLikeHandler.UnlikeComment)).ServeHTTP(w, r)
				return
			}
			if method == "GET" {
				authMW(http.HandlerFunc(jr.CommentLikeHandler.ListLikes)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- Routine Routes (with sub-resources) ---
	// GET /routines -> ListRoutines
	// POST /routines -> CreateRoutine
	// GET /routines/{id} -> GetRoutine
	// PUT /routines/{id} -> UpdateRoutine
	// DELETE /routines/{id} -> DeleteRoutine
	// POST /routines/{id}/exercises -> AddExercise (to routine)
	// DELETE /routines/{id}/exercises/{exerciseId} -> RemoveExercise (from routine)
	if strings.HasPrefix(path, "/routines") {
		parts := strings.Split(strings.Trim(path, "/"), "/")

		if len(parts) == 1 {
			// GET /routines or POST /routines
			if method == "GET" {
				authMW(http.HandlerFunc(jr.RoutineHandler.ListRoutines)).ServeHTTP(w, r)
				return
			}
			if method == "POST" {
				authMW(http.HandlerFunc(jr.RoutineHandler.CreateRoutine)).ServeHTTP(w, r)
				return
			}
		}

		if len(parts) == 2 {
			// GET/PUT/DELETE /routines/{id}
			if method == "GET" {
				authMW(http.HandlerFunc(jr.RoutineHandler.GetRoutine)).ServeHTTP(w, r)
				return
			}
			if method == "PUT" {
				authMW(http.HandlerFunc(jr.RoutineHandler.UpdateRoutine)).ServeHTTP(w, r)
				return
			}
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.RoutineHandler.DeleteRoutine)).ServeHTTP(w, r)
				return
			}
		}

		// /routines/{id}/exercises
		if len(parts) >= 3 && parts[2] == "exercises" {
			if len(parts) == 3 {
				// POST /routines/{id}/exercises -> AddExercise
				if method == "POST" {
					authMW(http.HandlerFunc(jr.RoutineExerciseHandler.AddExercise)).ServeHTTP(w, r)
					return
				}
			}
			if len(parts) == 4 {
				// DELETE /routines/{id}/exercises/{exerciseId} -> RemoveExercise
				if method == "DELETE" {
					authMW(http.HandlerFunc(jr.RoutineExerciseHandler.RemoveExercise)).ServeHTTP(w, r)
					return
				}
			}
		}
	}

	// --- Routine Exercises Sub-routes (for sets) ---
	// POST /routine-exercises/{id}/sets -> AddSet
	if strings.HasPrefix(path, "/routine-exercises") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		// /routine-exercises/{id}/sets
		if len(parts) == 3 && parts[2] == "sets" {
			if method == "POST" {
				authMW(http.HandlerFunc(jr.RoutineSetHandler.AddSet)).ServeHTTP(w, r)
				return
			}
		}
	}

	// --- Routine Sets Routes (standalone) ---
	// DELETE /routine-sets/{id} -> RemoveSet
	if strings.HasPrefix(path, "/routine-sets") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 2 {
			if method == "DELETE" {
				authMW(http.HandlerFunc(jr.RoutineSetHandler.RemoveSet)).ServeHTTP(w, r)
				return
			}
		}
	}

	// Default 404
	http.NotFound(w, r)
}

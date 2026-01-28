package testutil

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/rotsu1/jimu-backend/internal/handlers"
	"github.com/rotsu1/jimu-backend/internal/repository"
	router "github.com/rotsu1/jimu-backend/internal/routers"
)

// TestJWTSecret is the static secret used for signing JWTs in tests.
// Must match the value used by CreateTestToken.
const TestJWTSecret = "test-secret-key-123"

// TestServer holds the wired-up router and database pool for integration tests.
type TestServer struct {
	Router *router.JimuRouter
	DB     *pgxpool.Pool
}

// NewTestServer creates a fully-wired TestServer connected to the test database.
// It applies migrations and cleans up tables before returning.
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// 1. Get DATABASE_URL from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:pass@localhost:5432/jimu_test?sslmode=disable"
	}

	// 2. Apply migrations using standard sql.DB
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("Failed to open sql.DB for migrations: %v", err)
	}
	defer sqlDB.Close()

	migrations := &migrate.FileMigrationSource{
		Dir: "../../migrations",
	}

	n, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	if err != nil {
		t.Fatalf("Failed to apply migrations: %v", err)
	}
	t.Logf("Applied %d migrations", n)

	// 3. Create pgxpool for actual use
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to create pgxpool: %v", err)
	}

	// 4. Clean up tables for test isolation
	cleanupTables(t, pool)

	// 5. Initialize all Repositories (mirroring cmd/api/main.go)
	userRepo := repository.NewUserRepository(pool)
	userSessionRepo := repository.NewUserSessionRepository(pool)
	userDeviceRepo := repository.NewUserDeviceRepository(pool)
	subscriptionRepo := repository.NewSubscriptionRepository(pool)
	followRepo := repository.NewFollowRepository(pool)
	blockedUserRepo := repository.NewBlockedUserRepository(pool)
	workoutRepo := repository.NewWorkoutRepository(pool)
	workoutExerciseRepo := repository.NewWorkoutExerciseRepository(pool)
	workoutSetRepo := repository.NewWorkoutSetRepository(pool)
	workoutImageRepo := repository.NewWorkoutImageRepository(pool)
	workoutLikeRepo := repository.NewWorkoutLikeRepository(pool)
	exerciseRepo := repository.NewExerciseRepository(pool)
	muscleRepo := repository.NewMuscleRepository(pool)
	exerciseTargetMuscleRepo := repository.NewExerciseTargetMuscleRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)
	commentLikeRepo := repository.NewCommentLikeRepository(pool)
	routineRepo := repository.NewRoutineRepository(pool)
	routineExerciseRepo := repository.NewRoutineExerciseRepository(pool)
	routineSetRepo := repository.NewRoutineSetRepository(pool)

	// 6. Initialize all Handlers (mirroring cmd/api/main.go)
	authHandler := handlers.NewAuthHandler(userRepo, userSessionRepo, &handlers.GoogleValidator{})
	userSettingsHandler := handlers.NewUserSettingsHandler(userRepo)
	userDeviceHandler := handlers.NewUserDeviceHandler(userDeviceRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionRepo)
	followHandler := handlers.NewFollowHandler(followRepo)
	blockedUserHandler := handlers.NewBlockedUserHandler(blockedUserRepo)
	workoutHandler := handlers.NewWorkoutHandler(workoutRepo)
	workoutExerciseHandler := handlers.NewWorkoutExerciseHandler(workoutExerciseRepo)
	workoutSetHandler := handlers.NewWorkoutSetHandler(workoutSetRepo)
	workoutImageHandler := handlers.NewWorkoutImageHandler(workoutImageRepo)
	workoutLikeHandler := handlers.NewWorkoutLikeHandler(workoutLikeRepo)
	exerciseHandler := handlers.NewExerciseHandler(exerciseRepo)
	muscleHandler := handlers.NewMuscleHandler(muscleRepo)
	exerciseTargetMuscleHandler := handlers.NewExerciseTargetMuscleHandler(exerciseTargetMuscleRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo)
	commentLikeHandler := handlers.NewCommentLikeHandler(commentLikeRepo)
	routineHandler := handlers.NewRoutineHandler(routineRepo)
	routineExerciseHandler := handlers.NewRoutineExerciseHandler(routineExerciseRepo)
	routineSetHandler := handlers.NewRoutineSetHandler(routineSetRepo)

	// 7. Create Router (mirroring cmd/api/main.go)
	jimuRouter := &router.JimuRouter{
		AuthHandler:                 authHandler,
		UserSettingsHandler:         userSettingsHandler,
		UserDeviceHandler:           userDeviceHandler,
		SubscriptionHandler:         subscriptionHandler,
		FollowHandler:               followHandler,
		BlockedUserHandler:          blockedUserHandler,
		WorkoutHandler:              workoutHandler,
		WorkoutExerciseHandler:      workoutExerciseHandler,
		WorkoutSetHandler:           workoutSetHandler,
		WorkoutImageHandler:         workoutImageHandler,
		WorkoutLikeHandler:          workoutLikeHandler,
		ExerciseHandler:             exerciseHandler,
		MuscleHandler:               muscleHandler,
		ExerciseTargetMuscleHandler: exerciseTargetMuscleHandler,
		CommentHandler:              commentHandler,
		CommentLikeHandler:          commentLikeHandler,
		RoutineHandler:              routineHandler,
		RoutineExerciseHandler:      routineExerciseHandler,
		RoutineSetHandler:           routineSetHandler,
		JWTSecret:                   TestJWTSecret,
	}

	return &TestServer{
		Router: jimuRouter,
		DB:     pool,
	}
}

// cleanupTables truncates all test tables for isolation between tests.
func cleanupTables(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	const query = `
	TRUNCATE TABLE 
		public.routine_sets,
		public.routine_exercises,
		public.routines,
		public.comment_likes,
		public.comments,
		public.exercise_target_muscles,
		public.workout_likes,
		public.workout_images,
		public.workout_sets,
		public.workout_exercises,
		public.workouts,
		public.exercises,
		public.muscles,
		public.blocked_users,
		public.follows,
		public.subscriptions,
		public.user_devices,
		public.user_sessions,
		public.user_identities,
		public.profiles,
		public.sys_admins
	RESTART IDENTITY CASCADE`

	_, err := db.Exec(context.Background(), query)
	if err != nil {
		t.Fatalf("Failed to cleanup tables: %v", err)
	}
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rotsu1/jimu-backend/internal/db"
	"github.com/rotsu1/jimu-backend/internal/handlers"
	"github.com/rotsu1/jimu-backend/internal/repository"
	router "github.com/rotsu1/jimu-backend/internal/routers"
)

func main() {
	// 1. Initialize the DB Pool (from your internal/db package)
	pool, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer pool.Close()

	// 2. Initialize the Repository
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

	// 3. Initialize the Handler (Injecting the Repo)
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

	JWTSecret := os.Getenv("JWTSecret")
	if JWTSecret == "" {
		log.Fatal("JWTSecret is not set")
	}

	// 5. REGISTER THE HANDLER
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
		JWTSecret:                   JWTSecret,
	}

	// 6. Define the Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: jimuRouter,
	}
	// --- GRACEFUL SHUTDOWN LOGIC ---

	// Create a channel to listen for OS signals (Ctrl+C, Kill command)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the server in a separate goroutine so it doesn't block
	go func() {
		log.Println("Jimu Backend is starting on :8080... 🚀")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for the stop signal
	<-stop
	log.Println("\nShutdown signal received. Cleaning up... 🧹")

	// Create a deadline for the shutdown (e.g., 5 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the HTTP server first (stops accepting new requests)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close the Database pool
	log.Println("Closing database connections...")
	pool.Close()

	log.Println("Jimu Backend stopped gracefully. 気持ちいい！")
}

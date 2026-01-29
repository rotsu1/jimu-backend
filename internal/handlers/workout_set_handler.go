package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/middleware"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

type WorkoutSetScanner interface {
	CreateWorkoutSet(ctx context.Context, workoutExerciseID uuid.UUID, weight *float64, reps *int, isCompleted bool, orderIndex int, userID uuid.UUID) (*models.WorkoutSet, error)
	GetWorkoutSetByID(ctx context.Context, id uuid.UUID) (*models.WorkoutSet, error)
	GetWorkoutSetsByWorkoutExerciseID(ctx context.Context, workoutExerciseID uuid.UUID) ([]*models.WorkoutSet, error)
	UpdateWorkoutSet(ctx context.Context, workoutSetID uuid.UUID, userID uuid.UUID, updates models.UpdateWorkoutSetRequest) error
	DeleteWorkoutSet(ctx context.Context, workoutSetID uuid.UUID, userID uuid.UUID) error
}

type WorkoutSetHandler struct {
	Repo WorkoutSetScanner
}

func NewWorkoutSetHandler(r WorkoutSetScanner) *WorkoutSetHandler {
	return &WorkoutSetHandler{Repo: r}
}

func (h *WorkoutSetHandler) AddSet(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 2. Request Decoding
	// Workout Exercise ID from path or query?
	// Assuming query param "workout_exercise_id"
	workoutExerciseIDStr := r.URL.Query().Get("workout_exercise_id")
	if workoutExerciseIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		// /workout-exercises/WEID/sets
		if len(parts) >= 4 && parts[1] == "workout-exercises" && parts[3] == "sets" {
			workoutExerciseIDStr = parts[2]
		}
	}
	if workoutExerciseIDStr == "" {
		http.Error(w, "Missing workout exercise ID", http.StatusBadRequest)
		return
	}
	workoutExerciseID, err := uuid.Parse(workoutExerciseIDStr)
	if err != nil {
		http.Error(w, "Invalid workout exercise ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Weight      *float64 `json:"weight"`
		Reps        *int     `json:"reps"`
		IsCompleted bool     `json:"is_completed"`
		OrderIndex  int      `json:"order_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	ws, err := h.Repo.CreateWorkoutSet(r.Context(), workoutExerciseID, req.Weight, req.Reps, req.IsCompleted, req.OrderIndex, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Workout exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Add workout set error: %v", err)
		http.Error(w, "Failed to add set to workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ws)
}

func (h *WorkoutSetHandler) RemoveSet(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 2. Request Decoding
	targetIDStr := r.URL.Query().Get("id")
	if targetIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		// /workout-sets/ID
		if len(parts) >= 3 && parts[1] == "workout-sets" {
			targetIDStr = parts[2]
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing set ID", http.StatusBadRequest)
		return
	}
	setID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.DeleteWorkoutSet(r.Context(), setID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutSetNotFound) {
			http.Error(w, "Workout set not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete workout set error: %v", err)
		http.Error(w, "Failed to delete workout set", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutSetHandler) UpdateSet(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 2. Request Decoding
	targetIDStr := r.URL.Query().Get("id")
	if targetIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 3 && parts[1] == "workout-sets" {
			targetIDStr = parts[2]
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing set ID", http.StatusBadRequest)
		return
	}
	setID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateWorkoutSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.UpdateWorkoutSet(r.Context(), setID, userID, req)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutSetNotFound) {
			http.Error(w, "Workout set not found", http.StatusNotFound)
			return
		}
		log.Printf("Update workout set error: %v", err)
		http.Error(w, "Failed to update workout set", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/middleware"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

type WorkoutExerciseScanner interface {
	CreateWorkoutExercise(ctx context.Context, workoutID uuid.UUID, exerciseID uuid.UUID, orderIndex int, memo *string, restTimerSeconds *int, userID uuid.UUID) (*models.WorkoutExercise, error)
	GetWorkoutExerciseByID(ctx context.Context, id uuid.UUID) (*models.WorkoutExercise, error)
	GetWorkoutExercisesByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]*models.WorkoutExercise, error)
	UpdateWorkoutExercise(ctx context.Context, workoutExerciseID uuid.UUID, updates models.UpdateWorkoutExerciseRequest, userID uuid.UUID) error
	DeleteWorkoutExercise(ctx context.Context, workoutExerciseID uuid.UUID, userID uuid.UUID) error
}

type WorkoutExerciseHandler struct {
	Repo WorkoutExerciseScanner
}

func NewWorkoutExerciseHandler(r WorkoutExerciseScanner) *WorkoutExerciseHandler {
	return &WorkoutExerciseHandler{Repo: r}
}

func (h *WorkoutExerciseHandler) AddExercise(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/exercises)
	workoutID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing workout ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ExerciseID       string  `json:"exercise_id"`
		OrderIndex       int     `json:"order_index"`
		Memo             *string `json:"memo"`
		RestTimerSeconds *int    `json:"rest_timer_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	exerciseID, err := uuid.Parse(req.ExerciseID)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	we, err := h.Repo.CreateWorkoutExercise(r.Context(), workoutID, exerciseID, req.OrderIndex, req.Memo, req.RestTimerSeconds, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Workout or Exercise not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			http.Error(w, "Exercise already in workout", http.StatusConflict)
			return
		}
		log.Printf("Add workout exercise error: %v", err)
		http.Error(w, "Failed to add exercise to workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(we)
}

func (h *WorkoutExerciseHandler) RemoveExercise(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/exercises/{exerciseId})
	workoutExerciseID, err := GetUUIDPathParam(r, 3)
	if err != nil {
		http.Error(w, "Invalid or missing workout exercise ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.DeleteWorkoutExercise(r.Context(), workoutExerciseID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutExerciseNotFound) {
			http.Error(w, "Workout exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete workout exercise error: %v", err)
		http.Error(w, "Failed to remove exercise from workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutExerciseHandler) UpdateExercise(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/exercises/{exerciseId})
	workoutExerciseID, err := GetUUIDPathParam(r, 3)
	if err != nil {
		http.Error(w, "Invalid or missing workout exercise ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateWorkoutExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.UpdateWorkoutExercise(r.Context(), workoutExerciseID, req, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutExerciseNotFound) {
			http.Error(w, "Workout exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Update workout exercise error: %v", err)
		http.Error(w, "Failed to update workout exercise", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

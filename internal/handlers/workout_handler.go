package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/middleware"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

type WorkoutScanner interface {
	Create(ctx context.Context, userID uuid.UUID, name *string, comment *string, startedAt time.Time, endedAt time.Time, durationSeconds int) (*models.Workout, error)
	GetWorkoutByID(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID) (*models.Workout, error)
	GetWorkoutsByUserID(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.Workout, error)
	UpdateWorkout(ctx context.Context, id uuid.UUID, updates models.UpdateWorkoutRequest, userID uuid.UUID) error
	DeleteWorkout(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type WorkoutHandler struct {
	Repo WorkoutScanner
}

func NewWorkoutHandler(r WorkoutScanner) *WorkoutHandler {
	return &WorkoutHandler{Repo: r}
}

func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		Name            *string   `json:"name"`
		Comment         *string   `json:"comment"`
		StartedAt       time.Time `json:"started_at"`
		EndedAt         time.Time `json:"ended_at"`
		DurationSeconds int       `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	workout, err := h.Repo.Create(r.Context(), userID, req.Name, req.Comment, req.StartedAt, req.EndedAt, req.DurationSeconds)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Create workout error: %v", err)
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workout)
}

func (h *WorkoutHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
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
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			if _, err := uuid.Parse(lastPart); err == nil {
				targetIDStr = lastPart
			}
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}
	workoutID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	workout, err := h.Repo.GetWorkoutByID(r.Context(), workoutID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutNotFound) {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}
		log.Printf("Get workout error: %v", err)
		http.Error(w, "Failed to get workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workout)
}

func (h *WorkoutHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
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

	// 2. Determine target user (defaults to self)
	targetUserStr := r.URL.Query().Get("user_id")
	targetID := userID
	if targetUserStr != "" {
		parsed, err := uuid.Parse(targetUserStr)
		if err == nil {
			targetID = parsed
		}
	}

	limit := 20 // TODO: From query
	offset := 0

	// 3. Repo Call
	workouts, err := h.Repo.GetWorkoutsByUserID(r.Context(), targetID, userID, limit, offset)

	// 4. Error Mapping
	if err != nil {
		log.Printf("List workouts error: %v", err)
		http.Error(w, "Failed to list workouts", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if workouts == nil {
		workouts = []*models.Workout{}
	}
	json.NewEncoder(w).Encode(workouts)
}

func (h *WorkoutHandler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
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
	workoutID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing workout ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.UpdateWorkout(r.Context(), workoutID, req, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutNotFound) {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}
		log.Printf("Update workout error: %v", err)
		http.Error(w, "Failed to update workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutHandler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
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
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			if _, err := uuid.Parse(lastPart); err == nil {
				targetIDStr = lastPart
			}
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}
	workoutID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.DeleteWorkout(r.Context(), workoutID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutNotFound) {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete workout error: %v", err)
		http.Error(w, "Failed to delete workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

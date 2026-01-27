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

type ExerciseTargetMuscleScanner interface {
	GetByExerciseID(ctx context.Context, exerciseID uuid.UUID) ([]*models.ExerciseTargetMuscle, error)
	AddTargetMuscle(ctx context.Context, exerciseID uuid.UUID, muscleID uuid.UUID, userID uuid.UUID) (*models.ExerciseTargetMuscle, error)
	RemoveTargetMuscle(ctx context.Context, exerciseID uuid.UUID, muscleID uuid.UUID, userID uuid.UUID) error
}

type ExerciseTargetMuscleHandler struct {
	Repo ExerciseTargetMuscleScanner
}

func NewExerciseTargetMuscleHandler(r ExerciseTargetMuscleScanner) *ExerciseTargetMuscleHandler {
	return &ExerciseTargetMuscleHandler{Repo: r}
}

func (h *ExerciseTargetMuscleHandler) AddTargetMuscle(w http.ResponseWriter, r *http.Request) {
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
	// Exercise ID from path (or query for simplicity unless router used)
	exerciseIDStr := r.URL.Query().Get("id")
	if exerciseIDStr == "" {
		// Attempt parsing from path /exercises/{id}/muscles
		// Simple splitter
		parts := strings.Split(r.URL.Path, "/")
		// Expect /exercises/ID/muscles
		// parts: [ "", "exercises", "ID", "muscles" ]
		if len(parts) >= 4 && parts[1] == "exercises" && parts[3] == "muscles" {
			exerciseIDStr = parts[2]
		}
	}
	if exerciseIDStr == "" {
		http.Error(w, "Missing exercise ID", http.StatusBadRequest)
		return
	}
	exerciseID, err := uuid.Parse(exerciseIDStr)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	// Body: MuscleID
	var req struct {
		MuscleID string `json:"muscle_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	muscleID, err := uuid.Parse(req.MuscleID)
	if err != nil {
		http.Error(w, "Invalid muscle ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	etm, err := h.Repo.AddTargetMuscle(r.Context(), exerciseID, muscleID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Exercise or muscle not found", http.StatusNotFound)
			return
		}
		log.Printf("Add target muscle error: %v", err)
		http.Error(w, "Failed to add target muscle", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(etm)
}

func (h *ExerciseTargetMuscleHandler) RemoveTargetMuscle(w http.ResponseWriter, r *http.Request) {
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
	// Exercise ID and Muscle ID
	// Path: /exercises/{id}/muscles/{muscle_id}
	// Query: id=...&muscle_id=...
	exerciseIDStr := r.URL.Query().Get("id")
	muscleIDStr := r.URL.Query().Get("muscle_id")

	if exerciseIDStr == "" {
		// Try parsing path
		// /exercises/EID/muscles/MID
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 5 && parts[1] == "exercises" && parts[3] == "muscles" {
			exerciseIDStr = parts[2]
			muscleIDStr = parts[4]
		}
	}

	if exerciseIDStr == "" || muscleIDStr == "" {
		http.Error(w, "Missing IDs", http.StatusBadRequest)
		return
	}

	exerciseID, err := uuid.Parse(exerciseIDStr)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}
	muscleID, err := uuid.Parse(muscleIDStr)
	if err != nil {
		http.Error(w, "Invalid muscle ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.RemoveTargetMuscle(r.Context(), exerciseID, muscleID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Target muscle not found", http.StatusNotFound)
			return
		}
		log.Printf("Remove target muscle error: %v", err)
		http.Error(w, "Failed to remove target muscle", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

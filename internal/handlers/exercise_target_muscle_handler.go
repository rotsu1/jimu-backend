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
	// Path param only: /exercises/{id}/muscles
	exerciseID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing exercise ID", http.StatusBadRequest)
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
	// Path params only: /exercises/{id}/muscles/{muscleId}
	exerciseID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing exercise ID", http.StatusBadRequest)
		return
	}
	muscleID, err := GetUUIDPathParam(r, 3)
	if err != nil {
		http.Error(w, "Invalid or missing muscle ID", http.StatusBadRequest)
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

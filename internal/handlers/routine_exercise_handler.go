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

type RoutineExerciseScanner interface {
	CreateRoutineExercise(ctx context.Context, routineID uuid.UUID, exerciseID uuid.UUID, orderIndex int, restTimerSeconds *int, memo *string, userID uuid.UUID) (*models.RoutineExercise, error)
	GetRoutineExerciseByID(ctx context.Context, id uuid.UUID) (*models.RoutineExercise, error)
	GetRoutineExercisesByRoutineID(ctx context.Context, routineID uuid.UUID) ([]*models.RoutineExercise, error)
	UpdateRoutineExercise(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineExerciseRequest, userID uuid.UUID) error
	DeleteRoutineExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type RoutineExerciseHandler struct {
	Repo RoutineExerciseScanner
}

func NewRoutineExerciseHandler(r RoutineExerciseScanner) *RoutineExerciseHandler {
	return &RoutineExerciseHandler{Repo: r}
}

func (h *RoutineExerciseHandler) AddExercise(w http.ResponseWriter, r *http.Request) {
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
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	// The Routine ID is at index 2
	routineID, err := uuid.Parse(parts[2])
	if err != nil {
		http.Error(w, "Invalid or missing routine ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ExerciseID       string  `json:"exercise_id"`
		OrderIndex       int     `json:"order_index"`
		RestTimerSeconds *int    `json:"rest_timer_seconds"`
		Memo             *string `json:"memo"`
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
	re, err := h.Repo.CreateRoutineExercise(r.Context(), routineID, exerciseID, req.OrderIndex, req.RestTimerSeconds, req.Memo, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Routine or Exercise not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			http.Error(w, "Exercise already in routine", http.StatusConflict) // Unlikely if orderIndex allows dupes? Schema says unique? Check migration if needed, but error mapping handles it.
			return
		}
		log.Printf("Add routine exercise error: %v", err)
		http.Error(w, "Failed to add exercise to routine", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(re)
}

func (h *RoutineExerciseHandler) RemoveExercise(w http.ResponseWriter, r *http.Request) {
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
	routineExerciseID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing routine exercise ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.DeleteRoutineExercise(r.Context(), routineExerciseID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrRoutineExerciseNotFound) {
			http.Error(w, "Routine exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete routine exercise error: %v", err)
		http.Error(w, "Failed to remove exercise from routine", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

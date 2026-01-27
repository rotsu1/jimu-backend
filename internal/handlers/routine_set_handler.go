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

type RoutineSetScanner interface {
	CreateRoutineSet(ctx context.Context, routineExerciseID uuid.UUID, weight *float64, reps *int, orderIndex *int, userID uuid.UUID) (*models.RoutineSet, error)
	GetRoutineSetByID(ctx context.Context, id uuid.UUID) (*models.RoutineSet, error)
	GetRoutineSetsByRoutineExerciseID(ctx context.Context, routineExerciseID uuid.UUID) ([]*models.RoutineSet, error)
	UpdateRoutineSet(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineSetRequest, userID uuid.UUID) error
	DeleteRoutineSet(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type RoutineSetHandler struct {
	Repo RoutineSetScanner
}

func NewRoutineSetHandler(r RoutineSetScanner) *RoutineSetHandler {
	return &RoutineSetHandler{Repo: r}
}

func (h *RoutineSetHandler) AddSet(w http.ResponseWriter, r *http.Request) {
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
	// We need Routine Exercise ID.
	// Path: /routine-exercises/{id}/sets ?
	// Or /routines/{id}/exercises/{eid}/sets ?
	// Assuming /routine-exercises/{id}/sets for simplicitly or nested?
	// Based on Plan/Previous handlers, if we deleted exercise via /routines/{id}/exercises/{eid}, maybe sets are similar?
	// But `routine_id` isn't strictly needed for creating a set if we have `routine_exercise_id`.
	// Let's rely on query param `routine_exercise_id` if path is tricky or path parsing if structured.
	// Example: POST /routine-exercises/{id}/sets

	routineExerciseIDStr := r.URL.Query().Get("routine_exercise_id")
	if routineExerciseIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		// Expect /routine-exercises/ID/sets
		if len(parts) >= 4 && parts[1] == "routine-exercises" && parts[3] == "sets" {
			routineExerciseIDStr = parts[2]
		}
	}
	if routineExerciseIDStr == "" {
		http.Error(w, "Missing routine exercise ID", http.StatusBadRequest)
		return
	}
	routineExerciseID, err := uuid.Parse(routineExerciseIDStr)
	if err != nil {
		http.Error(w, "Invalid routine exercise ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Weight     *float64 `json:"weight"`
		Reps       *int     `json:"reps"`
		OrderIndex *int     `json:"order_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	rs, err := h.Repo.CreateRoutineSet(r.Context(), routineExerciseID, req.Weight, req.Reps, req.OrderIndex, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Routine exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Add routine set error: %v", err)
		http.Error(w, "Failed to add set to routine exercise", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rs)
}

func (h *RoutineSetHandler) RemoveSet(w http.ResponseWriter, r *http.Request) {
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
	// Set ID
	targetIDStr := r.URL.Query().Get("id")
	if targetIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		// Expect /routine-sets/ID
		// or /routine-exercises/EID/sets/SID
		if len(parts) >= 3 && parts[1] == "routine-sets" {
			targetIDStr = parts[2]
		} else if len(parts) >= 5 && parts[3] == "sets" {
			targetIDStr = parts[4] // /.../sets/ID
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
	err = h.Repo.DeleteRoutineSet(r.Context(), setID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrRoutineSetNotFound) {
			http.Error(w, "Routine set not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete routine set error: %v", err)
		http.Error(w, "Failed to delete routine set", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

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

type RoutineSetScanner interface {
	CreateRoutineSet(ctx context.Context, routineExerciseID uuid.UUID, weight *float64, reps *int, orderIndex int, userID uuid.UUID) (*models.RoutineSet, error)
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
	// Path param only: /routine-exercises/{id}/sets
	routineExerciseID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing routine exercise ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Weight     *float64 `json:"weight"`
		Reps       *int     `json:"reps"`
		OrderIndex int      `json:"order_index"`
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
	// Path param only: /routine-sets/{id}
	setID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing set ID", http.StatusBadRequest)
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

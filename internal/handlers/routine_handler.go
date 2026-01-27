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

type RoutineScanner interface {
	CreateRoutine(ctx context.Context, userID uuid.UUID, name string) (*models.Routine, error)
	GetRoutineByID(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) (*models.Routine, error)
	GetRoutinesByUserID(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID) ([]*models.Routine, error)
	UpdateRoutine(ctx context.Context, id uuid.UUID, updates models.UpdateRoutineRequest, userID uuid.UUID) error
	DeleteRoutine(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type RoutineHandler struct {
	Repo RoutineScanner
}

func NewRoutineHandler(r RoutineScanner) *RoutineHandler {
	return &RoutineHandler{Repo: r}
}

func (h *RoutineHandler) CreateRoutine(w http.ResponseWriter, r *http.Request) {
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
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	routine, err := h.Repo.CreateRoutine(r.Context(), userID, req.Name)

	// 4. Error Mapping
	if err != nil {
		// Log specific errors if any?
		// ErrReferenceViolation if user doesn't exist? (Unlikely due to auth check, but possible)
		log.Printf("Create routine error: %v", err)
		http.Error(w, "Failed to create routine", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(routine)
}

func (h *RoutineHandler) GetRoutine(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Missing routine ID", http.StatusBadRequest)
		return
	}
	routineID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid routine ID", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	routine, err := h.Repo.GetRoutineByID(r.Context(), routineID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrRoutineNotFound) {
			http.Error(w, "Routine not found", http.StatusNotFound)
			return
		}
		log.Printf("Get routine error: %v", err)
		http.Error(w, "Failed to get routine", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routine)
}

func (h *RoutineHandler) ListRoutines(w http.ResponseWriter, r *http.Request) {
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

	// 3. Repository Call
	routines, err := h.Repo.GetRoutinesByUserID(r.Context(), targetID, userID)

	// 4. Error Mapping
	if err != nil {
		log.Printf("List routines error: %v", err)
		http.Error(w, "Failed to list routines", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if routines == nil {
		routines = []*models.Routine{}
	}
	json.NewEncoder(w).Encode(routines)
}

func (h *RoutineHandler) UpdateRoutine(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Missing routine ID", http.StatusBadRequest)
		return
	}
	routineID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid routine ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRoutineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	err = h.Repo.UpdateRoutine(r.Context(), routineID, req, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrRoutineNotFound) {
			http.Error(w, "Routine not found", http.StatusNotFound)
			return
		}
		log.Printf("Update routine error: %v", err)
		http.Error(w, "Failed to update routine", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *RoutineHandler) DeleteRoutine(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Missing routine ID", http.StatusBadRequest)
		return
	}
	routineID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid routine ID", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	err = h.Repo.DeleteRoutine(r.Context(), routineID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrRoutineNotFound) {
			http.Error(w, "Routine not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete routine error: %v", err)
		http.Error(w, "Failed to delete routine", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

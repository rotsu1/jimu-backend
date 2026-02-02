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

type ExerciseScanner interface {
	CreateExercise(ctx context.Context, userID *uuid.UUID, name string, suggestedRestSeconds *int, icon *string, requesterID uuid.UUID) (*models.Exercise, error)
	GetExerciseByID(ctx context.Context, exerciseID uuid.UUID, userID uuid.UUID) (*models.Exercise, error)
	GetExercisesByUserID(ctx context.Context, viewerID uuid.UUID, targetID uuid.UUID) ([]*models.Exercise, error)
	UpdateExercise(ctx context.Context, id uuid.UUID, updates models.UpdateExerciseRequest, userID uuid.UUID) error
	DeleteExercise(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type ExerciseHandler struct {
	Repo ExerciseScanner
}

func NewExerciseHandler(r ExerciseScanner) *ExerciseHandler {
	return &ExerciseHandler{Repo: r}
}

func (h *ExerciseHandler) CreateExercise(w http.ResponseWriter, r *http.Request) {
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
		UserID               *string `json:"user_id"` // Optional, nil means system/admin created? Or if creating for self, usually implied.
		Name                 string  `json:"name"`
		SuggestedRestSeconds *int    `json:"suggested_rest_seconds"`
		Icon                 *string `json:"icon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var targetUserID *uuid.UUID
	if req.UserID != nil {
		parsedUID, err := uuid.Parse(*req.UserID)
		if err != nil {
			http.Error(w, "Invalid target user ID", http.StatusBadRequest)
			return
		}
		targetUserID = &parsedUID
	} else {
		// Default to creating for self
		targetUserID = &userID
	}

	// 3. Repository Call
	exercise, err := h.Repo.CreateExercise(r.Context(), targetUserID, req.Name, req.SuggestedRestSeconds, req.Icon, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			http.Error(w, "Exercise already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Invalid reference", http.StatusBadRequest)
			return
		}
		log.Printf("Create exercise error: %v", err)
		http.Error(w, "Failed to create exercise", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(exercise)
}

func (h *ExerciseHandler) GetExercise(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction
	// Path param only: /exercises/{id}
	exerciseID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing exercise ID", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	exercise, err := h.Repo.GetExerciseByID(r.Context(), exerciseID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrExerciseNotFound) {
			http.Error(w, "Exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Get exercise error: %v", err)
		http.Error(w, "Failed to get exercise", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercise)
}

func (h *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
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

	// 2. Determine whose exercises to fetch
	targetID := userID // Default to self
	targetUserStr := r.URL.Query().Get("user_id")
	if targetUserStr != "" {
		parsed, err := uuid.Parse(targetUserStr)
		if err == nil {
			targetID = parsed
		}
	}

	// 3. Repository Call
	exercises, err := h.Repo.GetExercisesByUserID(r.Context(), userID, targetID)
	if err != nil {
		log.Printf("List exercises error: %v", err)
		http.Error(w, "Failed to list exercises", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if exercises == nil {
		exercises = []*models.Exercise{}
	}
	json.NewEncoder(w).Encode(exercises)
}

func (h *ExerciseHandler) UpdateExercise(w http.ResponseWriter, r *http.Request) {
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
	exerciseID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing exercise ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	err = h.Repo.UpdateExercise(r.Context(), exerciseID, req, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrExerciseNotFound) {
			http.Error(w, "Exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Update exercise error: %v", err)
		http.Error(w, "Failed to update exercise", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *ExerciseHandler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
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
	exerciseID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing exercise ID", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	err = h.Repo.DeleteExercise(r.Context(), exerciseID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrExerciseNotFound) {
			http.Error(w, "Exercise not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete exercise error: %v", err)
		http.Error(w, "Failed to delete exercise", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

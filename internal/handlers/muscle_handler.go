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

type MuscleScanner interface {
	GetAllMuscles(ctx context.Context) ([]*models.Muscle, error)
	GetMuscleByID(ctx context.Context, id uuid.UUID) (*models.Muscle, error)
	GetMuscleByName(ctx context.Context, name string) (*models.Muscle, error)
	CreateMuscle(ctx context.Context, name string, userID uuid.UUID) (*models.Muscle, error)
	DeleteMuscle(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type MuscleHandler struct {
	Repo MuscleScanner
}

func NewMuscleHandler(r MuscleScanner) *MuscleHandler {
	return &MuscleHandler{Repo: r}
}

func (h *MuscleHandler) ListMuscles(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check (Optional for reading public data? But "Maze" says extract userID)
	// Assuming strictly authenticated for consistency.
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	_, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 2. Repo Call
	muscles, err := h.Repo.GetAllMuscles(r.Context())

	// 4. Error Mapping
	if err != nil {
		log.Printf("List muscles error: %v", err)
		http.Error(w, "Failed to list muscles", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if muscles == nil {
		muscles = []*models.Muscle{}
	}
	json.NewEncoder(w).Encode(muscles)
}

func (h *MuscleHandler) GetMuscle(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	_, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 2. Request Decoding
	// Path param only: /muscles/{id}
	muscleID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing muscle ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	muscle, err := h.Repo.GetMuscleByID(r.Context(), muscleID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrMuscleNotFound) {
			http.Error(w, "Muscle not found", http.StatusNotFound)
			return
		}
		log.Printf("Get muscle error: %v", err)
		http.Error(w, "Failed to get muscle", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(muscle)
}

func (h *MuscleHandler) CreateMuscle(w http.ResponseWriter, r *http.Request) {
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

	// 3. Repo Call
	muscle, err := h.Repo.CreateMuscle(r.Context(), req.Name, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			http.Error(w, "Muscle already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, repository.ErrUnauthorizedAction) {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
		log.Printf("Create muscle error: %v", err)
		http.Error(w, "Failed to create muscle", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(muscle)
}

func (h *MuscleHandler) DeleteMuscle(w http.ResponseWriter, r *http.Request) {
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
	muscleID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing muscle ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.DeleteMuscle(r.Context(), muscleID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrUnauthorizedAction) {
			http.Error(w, "Unauthorized or muscle not found", http.StatusForbidden)
			return
		}
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Muscle is in use", http.StatusConflict)
			return
		}
		log.Printf("Delete muscle error: %v", err)
		http.Error(w, "Failed to delete muscle", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

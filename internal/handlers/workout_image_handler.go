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

type WorkoutImageScanner interface {
	CreateWorkoutImage(ctx context.Context, workoutID uuid.UUID, storagePath string, displayOrder int, userID uuid.UUID) (*models.WorkoutImage, error)
	GetWorkoutImageByID(ctx context.Context, id uuid.UUID) (*models.WorkoutImage, error)
	GetWorkoutImagesByWorkoutID(ctx context.Context, workoutID uuid.UUID) ([]*models.WorkoutImage, error)
	DeleteWorkoutImage(ctx context.Context, workoutImageID uuid.UUID, userID uuid.UUID) error
}

type WorkoutImageHandler struct {
	Repo WorkoutImageScanner
}

func NewWorkoutImageHandler(r WorkoutImageScanner) *WorkoutImageHandler {
	return &WorkoutImageHandler{Repo: r}
}

func (h *WorkoutImageHandler) AddImage(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/images)
	workoutID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing workout ID", http.StatusBadRequest)
		return
	}

	var req struct {
		StoragePath  string `json:"storage_path"`
		DisplayOrder int    `json:"display_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.StoragePath == "" {
		http.Error(w, "Storage path required", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	wi, err := h.Repo.CreateWorkoutImage(r.Context(), workoutID, req.StoragePath, req.DisplayOrder, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			http.Error(w, "Image already exists", http.StatusConflict) // Unlikely unless path has unique constraint
			return
		}
		log.Printf("Add workout image error: %v", err)
		http.Error(w, "Failed to add image to workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wi)
}

func (h *WorkoutImageHandler) RemoveImage(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/images/{imageId})
	imageID, err := GetUUIDPathParam(r, 3)
	if err != nil {
		http.Error(w, "Invalid or missing image ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.DeleteWorkoutImage(r.Context(), imageID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutImageNotFound) {
			http.Error(w, "Workout image not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete workout image error: %v", err)
		http.Error(w, "Failed to delete workout image", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutImageHandler) ListImages(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/images)
	workoutID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing workout ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	images, err := h.Repo.GetWorkoutImagesByWorkoutID(r.Context(), workoutID)

	// 4. Error Mapping
	if err != nil {
		log.Printf("List workout images error: %v", err)
		http.Error(w, "Failed to list workout images", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if images == nil {
		images = []*models.WorkoutImage{}
	}
	json.NewEncoder(w).Encode(images)
}

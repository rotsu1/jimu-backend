package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/middleware"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

type WorkoutLikeScanner interface {
	LikeWorkout(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (*models.WorkoutLike, error)
	UnlikeWorkout(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) error
	GetWorkoutLikeByID(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (*models.WorkoutLike, error)
	GetWorkoutLikesByWorkoutID(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.WorkoutLikeDetail, error)
	IsWorkoutLiked(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID) (bool, error)
}

type WorkoutLikeHandler struct {
	Repo WorkoutLikeScanner
}

func NewWorkoutLikeHandler(r WorkoutLikeScanner) *WorkoutLikeHandler {
	return &WorkoutLikeHandler{Repo: r}
}

func (h *WorkoutLikeHandler) LikeWorkout(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/likes)
	workoutID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing workout ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	like, err := h.Repo.LikeWorkout(r.Context(), userID, workoutID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutInteractionNotAllowed) {
			http.Error(w, "Workout not found or blocked", http.StatusForbidden)
			return
		}
		log.Printf("Like workout error: %v", err)
		http.Error(w, "Failed to like workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(like)
}

func (h *WorkoutLikeHandler) UnlikeWorkout(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/likes)
	workoutID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing workout ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.UnlikeWorkout(r.Context(), userID, workoutID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutLikeNotFound) {
			http.Error(w, "Like not found", http.StatusNotFound)
			return
		}
		log.Printf("Unlike workout error: %v", err)
		http.Error(w, "Failed to unlike workout", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutLikeHandler) ListLikes(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /workouts/{id}/likes)
	workoutID, err := GetUUIDPathParam(r, 1)
	if err != nil {
		http.Error(w, "Invalid or missing workout ID", http.StatusBadRequest)
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// 3. Repo Call
	likes, err := h.Repo.GetWorkoutLikesByWorkoutID(r.Context(), workoutID, userID, limit, offset)

	// 4. Error Mapping
	if err != nil {
		log.Printf("List workout likes error: %v", err)
		http.Error(w, "Failed to list workout likes", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if likes == nil {
		likes = []*models.WorkoutLikeDetail{}
	}
	json.NewEncoder(w).Encode(likes)
}

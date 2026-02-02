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

type CommentScanner interface {
	CreateComment(ctx context.Context, userID uuid.UUID, workoutID uuid.UUID, parentID *uuid.UUID, content string) (*models.Comment, error)
	GetCommentByUserID(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) (*models.Comment, error)
	GetCommentsByWorkoutID(ctx context.Context, workoutID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.Comment, error)
	GetReplies(ctx context.Context, commentID uuid.UUID, viewerID uuid.UUID) ([]*models.Comment, error)
	DeleteComment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type CommentHandler struct {
	Repo CommentScanner
}

func NewCommentHandler(r CommentScanner) *CommentHandler {
	return &CommentHandler{Repo: r}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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
		WorkoutID string  `json:"workout_id"`
		ParentID  *string `json:"parent_id"` // For replies
		Content   string  `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	workoutID, err := uuid.Parse(req.WorkoutID)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			http.Error(w, "Invalid parent ID", http.StatusBadRequest)
			return
		}
		parentID = &pid
	}

	if req.Content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	comment, err := h.Repo.CreateComment(r.Context(), userID, workoutID, parentID, req.Content)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrWorkoutNotFound) {
			http.Error(w, "Workout not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Invalid reference", http.StatusBadRequest) // e.g. parent comment doesn't exist
			return
		}
		log.Printf("Create comment error: %v", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
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
	// Path param only: /comments/{id}
	commentID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing comment ID", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	comment, err := h.Repo.GetCommentByUserID(r.Context(), commentID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		log.Printf("Get comment error: %v", err)
		http.Error(w, "Failed to get comment", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
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

	// 2. ID Extraction (path param only: /comments/{id})
	commentID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing comment ID", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	err = h.Repo.DeleteComment(r.Context(), commentID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete comment error: %v", err)
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
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

	// 2. Query Params
	workoutIDStr := r.URL.Query().Get("workout_id")
	parentIDStr := r.URL.Query().Get("parent_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	var comments []*models.Comment

	// 3. Repo Call - Branch based on what we are listing
	if parentIDStr != "" {
		// FIX: Use 'parseErr' to avoid shadowing the main 'err' variable
		parentID, parseErr := uuid.Parse(parentIDStr)
		if parseErr != nil {
			http.Error(w, "Invalid parent ID", http.StatusBadRequest)
			return
		}
		// Now this correctly assigns to the outer 'err' and 'comments'
		comments, err = h.Repo.GetReplies(r.Context(), parentID, userID)

	} else if workoutIDStr != "" {
		// List Workout Comments
		// FIX: Use 'parseErr' here as well
		workoutID, parseErr := uuid.Parse(workoutIDStr)
		if parseErr != nil {
			http.Error(w, "Invalid workout ID", http.StatusBadRequest)
			return
		}
		comments, err = h.Repo.GetCommentsByWorkoutID(r.Context(), workoutID, userID, limit, offset)

	} else {
		http.Error(w, "Missing workout_id or parent_id", http.StatusBadRequest)
		return
	}

	// 4. Error Mapping
	// The linter error happened here because 'err' was always nil before the fix
	if err != nil {
		log.Printf("List comments error: %v", err)
		http.Error(w, "Failed to list comments", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if comments == nil {
		comments = []*models.Comment{}
	}
	json.NewEncoder(w).Encode(comments)
}

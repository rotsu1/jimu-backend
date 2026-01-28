package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/middleware"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

type CommentLikeScanner interface {
	LikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) (*models.CommentLike, error)
	UnlikeComment(ctx context.Context, userID uuid.UUID, commentID uuid.UUID) error
	GetCommentLikesByCommentID(ctx context.Context, commentID uuid.UUID, viewerID uuid.UUID, limit int, offset int) ([]*models.CommentLikeDetail, error)
}

type CommentLikeHandler struct {
	Repo CommentLikeScanner
}

func NewCommentLikeHandler(r CommentLikeScanner) *CommentLikeHandler {
	return &CommentLikeHandler{Repo: r}
}

func (h *CommentLikeHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
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

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	commentID, err := uuid.Parse(parts[len(parts)-2])
	if err != nil {
		http.Error(w, "Invalid or missing comment ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	like, err := h.Repo.LikeComment(r.Context(), userID, commentID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrCommentInteractionNotAllowed) {
			http.Error(w, "Interaction not allowed", http.StatusForbidden)
			return
		}
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		log.Printf("Like comment error: %v", err)
		http.Error(w, "Failed to like comment", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(like)
}

func (h *CommentLikeHandler) UnlikeComment(w http.ResponseWriter, r *http.Request) {
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

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	commentID, err := uuid.Parse(parts[len(parts)-2])
	if err != nil {
		http.Error(w, "Invalid or missing comment ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.UnlikeComment(r.Context(), userID, commentID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrCommentLikeNotFound) {
			http.Error(w, "Like not found", http.StatusNotFound)
			return
		}
		log.Printf("Unlike comment error: %v", err)
		http.Error(w, "Failed to unlike comment", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *CommentLikeHandler) ListLikes(w http.ResponseWriter, r *http.Request) {
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
	commentIDStr := r.URL.Query().Get("id")
	if commentIDStr == "" {
		http.Error(w, "Missing comment ID (query param 'id')", http.StatusBadRequest)
		return
	}
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

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

	// 3. Repo Call
	likes, err := h.Repo.GetCommentLikesByCommentID(r.Context(), commentID, userID, limit, offset)

	// 4. Error Mapping
	if err != nil {
		log.Printf("List comment likes error: %v", err)
		http.Error(w, "Failed to list likes", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if likes == nil {
		likes = []*models.CommentLikeDetail{}
	}
	json.NewEncoder(w).Encode(likes)
}

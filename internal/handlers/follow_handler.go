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

type FollowScanner interface {
	Follow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*models.Follow, error)
	Unfollow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error
	GetFollowStatus(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*models.Follow, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Follow, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]*models.Follow, error)
}

type FollowHandler struct {
	Repo FollowScanner
}

func NewFollowHandler(r FollowScanner) *FollowHandler {
	return &FollowHandler{Repo: r}
}

func (h *FollowHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	followerID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 2. Request Decoding
	// Path: /users/{id}/follow
	// Query: id=...
	targetIDStr := r.URL.Query().Get("id")
	if targetIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		// Expect /users/ID/follow
		if len(parts) >= 4 && parts[1] == "users" && parts[3] == "follow" {
			targetIDStr = parts[2]
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	followingID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if followerID == followingID {
		http.Error(w, "Cannot follow yourself", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	follow, err := h.Repo.Follow(r.Context(), followerID, followingID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, repository.ErrBlocked) {
			http.Error(w, "Blocked", http.StatusForbidden)
			return
		}
		log.Printf("Follow error: %v", err)
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(follow)
}

func (h *FollowHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	followerID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 2. Request Decoding
	targetIDStr := r.URL.Query().Get("id")
	if targetIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 && parts[1] == "users" && parts[3] == "follow" {
			targetIDStr = parts[2]
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	followingID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.Unfollow(r.Context(), followerID, followingID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrFollowNotFound) {
			http.Error(w, "Follow not found", http.StatusNotFound)
			return
		}
		log.Printf("Unfollow error: %v", err)
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *FollowHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
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
	targetIDStr := r.URL.Query().Get("id")
	if targetIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		// Expect /users/ID/followers
		if len(parts) >= 4 && parts[1] == "users" && parts[3] == "followers" {
			targetIDStr = parts[2]
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
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
	followers, err := h.Repo.GetFollowers(r.Context(), targetID, limit, offset)

	// 4. Error Mapping
	if err != nil {
		log.Printf("Get followers error: %v", err)
		http.Error(w, "Failed to get followers", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if followers == nil {
		followers = []*models.Follow{}
	}
	json.NewEncoder(w).Encode(followers)
}

func (h *FollowHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
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
	targetIDStr := r.URL.Query().Get("id")
	if targetIDStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		// Expect /users/ID/following
		if len(parts) >= 4 && parts[1] == "users" && parts[3] == "following" {
			targetIDStr = parts[2]
		}
	}
	if targetIDStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
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
	following, err := h.Repo.GetFollowing(r.Context(), targetID, limit, offset)

	// 4. Error Mapping
	if err != nil {
		log.Printf("Get following error: %v", err)
		http.Error(w, "Failed to get following", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if following == nil {
		following = []*models.Follow{}
	}
	json.NewEncoder(w).Encode(following)
}

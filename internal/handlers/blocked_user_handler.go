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

type BlockedUserScanner interface {
	Block(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) (*models.BlockedUser, error)
	Unblock(ctx context.Context, blockerID uuid.UUID, blockedID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]*models.BlockedUser, error)
}

type BlockedUserHandler struct {
	Repo BlockedUserScanner
}

func NewBlockedUserHandler(r BlockedUserScanner) *BlockedUserHandler {
	return &BlockedUserHandler{Repo: r}
}

func (h *BlockedUserHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
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
		BlockedID string `json:"blocked_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	blockedID, err := uuid.Parse(req.BlockedID)
	if err != nil {
		http.Error(w, "Invalid blocked_id format", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	_, err = h.Repo.Block(r.Context(), userID, blockedID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrReferenceViolation) {
			http.Error(w, "User does not exist", http.StatusNotFound)
			return
		}
		log.Printf("Block user error: %v", err)
		http.Error(w, "Failed to block user", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusCreated)
}

func (h *BlockedUserHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
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
	if len(parts) == 0 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	blockedID, err := uuid.Parse(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid or missing user ID", http.StatusBadRequest)
		return
	}

	// 3. Repository Call
	err = h.Repo.Unblock(r.Context(), userID, blockedID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrBlockedUserNotFound) {
			http.Error(w, "Blocked user not found", http.StatusNotFound)
			return
		}
		log.Printf("Unblock user error: %v", err)
		http.Error(w, "Failed to unblock user", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

func (h *BlockedUserHandler) GetBlockedUsers(w http.ResponseWriter, r *http.Request) {
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

	// 2. Repository Call
	users, err := h.Repo.GetBlockedUsers(r.Context(), userID)
	if err != nil {
		log.Printf("Get blocked users error: %v", err)
		http.Error(w, "Failed to fetch blocked users", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if users == nil {
		users = []*models.BlockedUser{}
	}
	json.NewEncoder(w).Encode(users)
}

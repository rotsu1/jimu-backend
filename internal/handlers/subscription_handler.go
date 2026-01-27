package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/middleware"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

type SubscriptionScanner interface {
	UpsertSubscription(ctx context.Context, userID uuid.UUID, originalTransactionID string, productID string, status string, expiresAt time.Time, environment string) (*models.Subscription, error)
	GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID) (*models.Subscription, error)
	GetSubscriptionByTransactionID(ctx context.Context, transactionID string, userID uuid.UUID) (*models.Subscription, error)
	DeleteSubscription(ctx context.Context, userID uuid.UUID) error
}

type SubscriptionHandler struct {
	Repo SubscriptionScanner
}

func NewSubscriptionHandler(r SubscriptionScanner) *SubscriptionHandler {
	return &SubscriptionHandler{Repo: r}
}

func (h *SubscriptionHandler) UpsertSubscription(w http.ResponseWriter, r *http.Request) {
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
		OriginalTransactionID string    `json:"original_transaction_id"`
		ProductID             string    `json:"product_id"`
		Status                string    `json:"status"`
		ExpiresAt             time.Time `json:"expires_at"`
		Environment           string    `json:"environment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	sub, err := h.Repo.UpsertSubscription(r.Context(), userID, req.OriginalTransactionID, req.ProductID, req.Status, req.ExpiresAt, req.Environment)

	// 4. Error Mapping
	if err != nil {
		log.Printf("Upsert subscription error: %v", err)
		http.Error(w, "Failed to upsert subscription", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusOK) // or Created? Upsert is mixed. 200 OK is safe.
	json.NewEncoder(w).Encode(sub)
}

func (h *SubscriptionHandler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
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

	// 2. Repo Call
	sub, err := h.Repo.GetSubscriptionByUserID(r.Context(), userID, userID)

	// 3. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}
		log.Printf("Get subscription error: %v", err)
		http.Error(w, "Failed to get subscription", http.StatusInternalServerError)
		return
	}

	// 4. Response Construction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

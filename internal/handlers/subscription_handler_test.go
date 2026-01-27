package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/handlers/testutils"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
)

// --- Mocks ---

type mockSubscriptionRepo struct {
	UpsertSubscriptionFunc             func(ctx context.Context, userID uuid.UUID, originalTransactionID string, productID string, status string, expiresAt time.Time, environment string) (*models.Subscription, error)
	GetSubscriptionByUserIDFunc        func(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID) (*models.Subscription, error)
	GetSubscriptionByTransactionIDFunc func(ctx context.Context, transactionID string, userID uuid.UUID) (*models.Subscription, error)
	DeleteSubscriptionFunc             func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockSubscriptionRepo) UpsertSubscription(ctx context.Context, userID uuid.UUID, originalTransactionID string, productID string, status string, expiresAt time.Time, environment string) (*models.Subscription, error) {
	if m.UpsertSubscriptionFunc != nil {
		return m.UpsertSubscriptionFunc(ctx, userID, originalTransactionID, productID, status, expiresAt, environment)
	}
	return &models.Subscription{UserID: userID, ProductID: productID}, nil
}

func (m *mockSubscriptionRepo) GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID) (*models.Subscription, error) {
	if m.GetSubscriptionByUserIDFunc != nil {
		return m.GetSubscriptionByUserIDFunc(ctx, userID, viewerID)
	}
	return &models.Subscription{UserID: userID}, nil
}

func (m *mockSubscriptionRepo) GetSubscriptionByTransactionID(ctx context.Context, transactionID string, userID uuid.UUID) (*models.Subscription, error) {
	if m.GetSubscriptionByTransactionIDFunc != nil {
		return m.GetSubscriptionByTransactionIDFunc(ctx, transactionID, userID)
	}
	return nil, nil
}

func (m *mockSubscriptionRepo) DeleteSubscription(ctx context.Context, userID uuid.UUID) error {
	if m.DeleteSubscriptionFunc != nil {
		return m.DeleteSubscriptionFunc(ctx, userID)
	}
	return nil
}

// --- Tests ---

func TestUpsertSubscription_Success(t *testing.T) {
	h := NewSubscriptionHandler(&mockSubscriptionRepo{})

	body := `{"product_id": "premium", "status": "active"}`
	req := httptest.NewRequest("POST", "/subscriptions", strings.NewReader(body))
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.UpsertSubscription(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestGetMySubscription_Success(t *testing.T) {
	h := NewSubscriptionHandler(&mockSubscriptionRepo{})

	req := httptest.NewRequest("GET", "/subscriptions/me", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetMySubscription(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestGetMySubscription_NotFound(t *testing.T) {
	mockRepo := &mockSubscriptionRepo{
		GetSubscriptionByUserIDFunc: func(ctx context.Context, userID uuid.UUID, viewerID uuid.UUID) (*models.Subscription, error) {
			return nil, repository.ErrSubscriptionNotFound
		},
	}
	h := NewSubscriptionHandler(mockRepo)

	req := httptest.NewRequest("GET", "/subscriptions/me", nil)
	req = testutils.InjectUserID(req, uuid.New().String())
	rr := httptest.NewRecorder()

	h.GetMySubscription(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", rr.Code)
	}
}

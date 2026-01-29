package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHealthRepo struct {
	PingFunc func(ctx context.Context) error
}

func (m *mockHealthRepo) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

func TestHealthCheck_Healthy(t *testing.T) {
	mockRepo := &mockHealthRepo{}
	h := NewHealthHandler(mockRepo)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	h.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if rr.Body.String() != "Healthy" {
		t.Errorf("expected Healthy, got %s", rr.Body.String())
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	mockRepo := &mockHealthRepo{
		PingFunc: func(ctx context.Context) error {
			return errors.New("db down")
		},
	}
	h := NewHealthHandler(mockRepo)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	h.HealthCheck(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 Service Unavailable, got %d", rr.Code)
	}
}

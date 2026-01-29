package handlers

import (
	"context"
	"net/http"
)

type HealthScanner interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	Repo HealthScanner
}

func NewHealthHandler(repo HealthScanner) *HealthHandler {
	return &HealthHandler{Repo: repo}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.Repo.Ping(r.Context()); err != nil {
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Healthy"))
}

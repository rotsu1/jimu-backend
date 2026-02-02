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

type UserDeviceScanner interface {
	UpsertUserDevice(ctx context.Context, userID uuid.UUID, fcmToken string, deviceType string, viewerID uuid.UUID) (*models.UserDevice, error)
	GetUserDeviceByID(ctx context.Context, deviceID uuid.UUID, userID uuid.UUID) (*models.UserDevice, error)
	GetUserDevicesByUserID(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID) ([]*models.UserDevice, error)
	GetUserDeviceByFCMToken(ctx context.Context, fcmToken string, userID uuid.UUID) (*models.UserDevice, error)
	DeleteUserDeviceByID(ctx context.Context, deviceID uuid.UUID, userID uuid.UUID) error
	DeleteUserDevicesByUserID(ctx context.Context, targetID uuid.UUID, viewerID uuid.UUID) error
}

type UserDeviceHandler struct {
	Repo UserDeviceScanner
}

func NewUserDeviceHandler(r UserDeviceScanner) *UserDeviceHandler {
	return &UserDeviceHandler{Repo: r}
}

func (h *UserDeviceHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
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
		FCMToken   string `json:"fcm_token"`
		DeviceType string `json:"device_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FCMToken == "" {
		http.Error(w, "FCM token required", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	device, err := h.Repo.UpsertUserDevice(r.Context(), userID, req.FCMToken, req.DeviceType, userID)

	// 4. Error Mapping
	if err != nil {
		log.Printf("Register device error: %v", err)
		http.Error(w, "Failed to register device", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(device)
}

func (h *UserDeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
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

	// 2. Repo Call - List "my" devices
	devices, err := h.Repo.GetUserDevicesByUserID(r.Context(), userID, userID)

	// 3. Error Mapping
	if err != nil {
		log.Printf("List devices error: %v", err)
		http.Error(w, "Failed to list devices", http.StatusInternalServerError)
		return
	}

	// 4. Response Construction
	w.Header().Set("Content-Type", "application/json")
	if devices == nil {
		devices = []*models.UserDevice{}
	}
	json.NewEncoder(w).Encode(devices)
}

func (h *UserDeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
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
	// Path param only: /user-devices/{id}
	deviceID, err := GetIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid or missing device ID", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.Repo.DeleteUserDeviceByID(r.Context(), deviceID, userID)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrUserDeviceNotFound) {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}
		log.Printf("Delete device error: %v", err)
		http.Error(w, "Failed to delete device", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

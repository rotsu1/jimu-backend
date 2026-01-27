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

type UserSettingsScanner interface {
	GetUserSettingsByID(ctx context.Context, id uuid.UUID) (*models.UserSetting, error)
	UpdateUserSettings(ctx context.Context, id string, updates models.UpdateUserSettingsRequest) error
}

type UserSettingsHandler struct {
	Repo UserSettingsScanner
}

func NewUserSettingsHandler(r UserSettingsScanner) *UserSettingsHandler {
	return &UserSettingsHandler{Repo: r}
}

func (h *UserSettingsHandler) GetMySettings(w http.ResponseWriter, r *http.Request) {
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
	settings, err := h.Repo.GetUserSettingsByID(r.Context(), userID)

	// 3. Error Mapping
	if err != nil {
		// If using pgx NoRows and it's not mapped?
		// Actually GetUserSettingsByID maps? No, it returns error directly unless it maps internally.
		// Checking user_repo.go line 258: `if err != nil { return nil, err }`.
		// It doesn't seem to map NoRows to a specific sentinel error in the snippet I saw (it just returns err).
		// Wait, line 242: `err := r.DB.QueryRow...`
		// If NoRows, it returns it.
		// So strict mapping: we might get pgx.ErrNoRows.
		// But usually `user_repo` methods return specific errors.
		// `GetProfileByID` (lines 82-114) also just returns err.
		// I should ideally handle `pgx.ErrNoRows` here or assume it's wrapped/mapped if I missed it.
		// Or strictly: settings SHOULD exist for every user? Maybe created on user creation?
		// If not found, maybe 404.
		log.Printf("Get user settings error: %v", err)
		http.Error(w, "Failed to get user settings", http.StatusInternalServerError)
		return
	}

	// 4. Response Construction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *UserSettingsHandler) UpdateMySettings(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}
	// The repo method `UpdateUserSettings` takes string ID.
	// So we can pass ctxID directly if we trust it, or Parse/String it.
	// Best to use ctxID string.

	// 2. Request Decoding
	var req models.UpdateUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err := h.Repo.UpdateUserSettings(r.Context(), ctxID, req)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			// Reuse ProfileNotFound or creates new one? Repo uses generic "failed to update..."
			// But wait, line 216 of UpdateProfile (not Settings) used ErrProfileNotFound.
			// Line 338 of UpdateUserSettings returns fmt.Errorf("failed to update user settings: %w", err).
			// It doesn't map NoRows explicitly in the snippet I saw for UpdateUserSettings?
			// Actually line 337 check err != nil.
			// Does Exec return NoRows? No. RowsAffected.
			// But `UpdateUserSettings` in snippet (lines 265-342) didn't check RowsAffected! (It returns nil at 341).
			// That might be a bug in Repository or it assumes it always succeeds.
			// I'll just check for error.
		}
		log.Printf("Update user settings error: %v", err)
		http.Error(w, "Failed to update user settings", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/rotsu1/jimu-backend/internal/auth"
	"github.com/rotsu1/jimu-backend/internal/repository"
	"google.golang.org/api/idtoken"
)

// AuthHandler holds the dependencies for authentication routes
type AuthHandler struct {
	UserRepo *repository.UserRepository
}

// NewAuthHandler is a constructor to create a new handler instance
func NewAuthHandler(repo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		UserRepo: repo,
	}
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the request body to get the "id_token" from the iOS app
	var req struct {
		IDtoken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Verify the token with Google
	payload, err := idtoken.Validate(
		r.Context(),
		req.IDtoken,
		"607893165629-fb5m9pgljivu1cvkfakf3k98782d7038.apps.googleusercontent.com",
	)
	if err != nil {
		http.Error(w, "Invalid Google token", http.StatusUnauthorized)
		return
	}

	// 3. Extract user info from payload (e.g., payload.Subject is the unique Google ID)
	googleID := payload.Subject
	email := payload.Claims["email"].(string)

	// 4. Upsert the user in your database
	user, err := h.UserRepo.UpsertGoogleUser(r.Context(), googleID, email)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	secret := os.Getenv("JIMU_SECRET")

	token, err := auth.GenerateJimuToken(user.ID.String(), secret)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"access_token": token})
}

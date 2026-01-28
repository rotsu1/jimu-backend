package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"

	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/auth"
	"github.com/rotsu1/jimu-backend/internal/middleware"
	"github.com/rotsu1/jimu-backend/internal/models"
	"github.com/rotsu1/jimu-backend/internal/repository"
	"google.golang.org/api/idtoken"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserScanner interface {
	UpsertGoogleUser(ctx context.Context, googleID, email string) (*models.Profile, error)
	GetProfileByID(ctx context.Context, viewerID uuid.UUID, targetID uuid.UUID) (*models.Profile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates models.UpdateProfileRequest) error
	DeleteProfile(ctx context.Context, id uuid.UUID) error
	GetIdentitiesByUserID(ctx context.Context, userID uuid.UUID) ([]*models.UserIdentity, error)
	DeleteIdentity(ctx context.Context, userID uuid.UUID, provider string) error
}

type SessionScanner interface {
	CreateSession(ctx context.Context, userID uuid.UUID, token string, agent *string, ip *string, exp time.Time) (*models.UserSession, error)
	GetSessionByRefreshToken(ctx context.Context, token string) (*models.UserSession, error)
	RevokeSession(ctx context.Context, id uuid.UUID, viewerID uuid.UUID) error
	RevokeAllSessionsForUser(ctx context.Context, targetUserID uuid.UUID, viewerID uuid.UUID) error
}

type TokenValidator interface {
	Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
}

// Real validator for production
type GoogleValidator struct{}

func (v *GoogleValidator) Validate(
	ctx context.Context,
	idToken string,
	audience string,
) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, idToken, audience)
}

type AuthHandler struct {
	UserRepo        UserScanner
	UserSessionRepo SessionScanner
	Validator       TokenValidator
}

// NewAuthHandler is a constructor to create a new handler instance
func NewAuthHandler(
	userRepo UserScanner,
	userSessionRepo SessionScanner,
	validator TokenValidator,
) *AuthHandler {
	return &AuthHandler{
		UserRepo:        userRepo,
		UserSessionRepo: userSessionRepo,
		Validator:       validator,
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
	payload, err := h.Validator.Validate(
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

	user, err := h.UserRepo.UpsertGoogleUser(r.Context(), googleID, email)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// 4. Generate the Dual-Token Pair
	secret := os.Getenv("JIMU_SECRET")
	accessToken, refreshToken, expiresIn, err := auth.GenerateTokenPair(user.ID.String(), secret)
	if err != nil {
		http.Error(w, "Token generation error", http.StatusInternalServerError)
		return
	}

	// 5. Create the Session in the Database
	// We set the expiry to match the Refresh Token (1 year)
	expiresAt := time.Now().Add(time.Hour * 24 * 365)
	userAgent := r.UserAgent()
	var ipPtr *string
	if r.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ipPtr = &host
		} else {
			// If no port, use the whole string
			ipPtr = &r.RemoteAddr
		}
	}

	_, err = h.UserSessionRepo.CreateSession(
		r.Context(),
		user.ID,
		refreshToken,
		&userAgent,
		ipPtr,
		expiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// 6. Return both tokens to the iOS app
	w.Header().Set("Content-Type", "application/json")

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the refresh token from the request
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 2. Look up the session in your database using your Repo
	// This checks if the token exists, isn't revoked, and isn't expired
	session, err := h.UserSessionRepo.GetSessionByRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		log.Printf("Session lookup failed: %v", err)
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	// 3. Generate a NEW token pair
	secret := os.Getenv("JIMU_SECRET")
	newAccessToken, newRefreshToken, expiresIn, err := auth.GenerateTokenPair(session.UserID.String(), secret)
	if err != nil {
		http.Error(w, "Token generation error", http.StatusInternalServerError)
		return
	}

	// 4. ROTATION: Revoke the old session and create the new one
	err = h.UserSessionRepo.RevokeSession(r.Context(), session.ID, session.UserID)
	if err != nil {
		log.Printf("Failed to revoke old session: %v", err)
	}

	userAgent := r.UserAgent()
	// Handle IP address (strip port if present, ignore errors)
	var ipPtr *string
	if r.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ipPtr = &host
		} else {
			ipPtr = &r.RemoteAddr
		}
	}
	expiresAt := time.Now().Add(time.Hour * 24 * 365)
	_, err = h.UserSessionRepo.CreateSession(
		r.Context(),
		session.UserID,
		newRefreshToken,
		&userAgent,
		ipPtr,
		expiresAt,
	)
	if err != nil {
		http.Error(w, "Failed to rotate session", http.StatusInternalServerError)
		return
	}

	// 5. Send the new pair back to the Swift app
	w.Header().Set("Content-Type", "application/json")

	response := AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 1. Get the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// 2. Parse the Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid auth header format", http.StatusUnauthorized)
		return
	}
	tokenString := parts[1]

	// 3. Verify the token to get the UserID (to ensure the user owns the session)
	secret := os.Getenv("JIMU_SECRET")
	userIDStr, err := auth.VerifyToken(tokenString, secret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID, _ := uuid.Parse(userIDStr)

	// 4. Revoke the session in the database
	err = h.UserSessionRepo.RevokeAllSessionsForUser(r.Context(), userID, userID)
	if err != nil {
		log.Printf("Logout failed: %v", err)
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// 5. Respond with success
	w.WriteHeader(http.StatusNoContent) // 204 No Content is standard for successful logout
}

func (h *AuthHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the userID from the Context
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// 2. Fetch the profile
	profile, err := h.UserRepo.GetProfileByID(r.Context(), userID, userID)
	if err != nil {
		log.Printf("Profile fetch error: %v", err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *AuthHandler) GetOtherProfile(w http.ResponseWriter, r *http.Request) {
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	viewerID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 0 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	targetID, err := uuid.Parse(parts[len(parts)-1])
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	profile, err := h.UserRepo.GetProfileByID(r.Context(), viewerID, targetID)
	if err != nil {
		log.Printf("Profile fetch error: %v", err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *AuthHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the userID from the Context
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// 2. Decode the request body to get the profile data
	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 3. Update the profile
	err = h.UserRepo.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		// Check for specific errors like 'Username Taken'
		if errors.Is(err, repository.ErrUsernameTaken) {
			http.Error(w, "Username already taken", http.StatusConflict)
			return
		}
		log.Printf("Profile update error: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// 4. Return Success
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) DeleteMyProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the userID from the Context
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// 2. Delete the profile
	err = h.UserRepo.DeleteProfile(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
		log.Printf("Profile deletion error: %v", err)
		http.Error(w, "Failed to delete profile", http.StatusInternalServerError)
		return
	}

	// 3. Return Success
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) GetMyIdentities(w http.ResponseWriter, r *http.Request) {
	// 1. Grab the userID from the Context
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// 2. Fetch the identities
	identities, err := h.UserRepo.GetIdentitiesByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Identities fetch error: %v", err)
		http.Error(w, "Failed to fetch identities", http.StatusInternalServerError)
		return
	}

	// 3. Return Success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(identities)
}

func (h *AuthHandler) UnlinkIdentity(w http.ResponseWriter, r *http.Request) {
	// 1. Context Check
	ctxID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(ctxID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// 2. Request Decoding
	var req struct {
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Provider == "" {
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}

	// 3. Repo Call
	err = h.UserRepo.DeleteIdentity(r.Context(), userID, req.Provider)

	// 4. Error Mapping
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Identity not found", http.StatusNotFound)
			return
		}
		log.Printf("Identity unlink error: %v", err)
		http.Error(w, "Failed to unlink identity", http.StatusInternalServerError)
		return
	}

	// 5. Response Construction
	w.WriteHeader(http.StatusNoContent)
}

package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/auth"
	"github.com/rotsu1/jimu-backend/internal/models"
	"google.golang.org/api/idtoken"
)

type UserScanner interface {
	UpsertGoogleUser(ctx context.Context, googleID, email string) (*models.Profile, error)
}

type SessionScanner interface {
	CreateSession(ctx context.Context, userID uuid.UUID, token string, agent *string, ip *netip.Addr, exp time.Time) (*models.UserSession, error)
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
	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID.String(), secret)
	if err != nil {
		http.Error(w, "Token generation error", http.StatusInternalServerError)
		return
	}

	// 5. Create the Session in the Database
	// We set the expiry to match the Refresh Token (1 year)
	expiresAt := time.Now().Add(time.Hour * 24 * 365)
	userAgent := r.UserAgent()
	var ipPtr *netip.Addr
	if r.RemoteAddr != "" {
		// RemoteAddr usually includes a port (e.g., "127.0.0.1:1234")
		// we need to strip it to get just the IP
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if parsedIP, err := netip.ParseAddr(host); err == nil {
				ipPtr = &parsedIP
			}
		} else {
			// If there's no port, just try parsing the whole string
			if parsedIP, err := netip.ParseAddr(r.RemoteAddr); err == nil {
				ipPtr = &parsedIP
			}
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
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
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
	newAccessToken, newRefreshToken, err := auth.GenerateTokenPair(session.UserID.String(), secret)
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
	var ipPtr *netip.Addr
	if r.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			if parsedIP, err := netip.ParseAddr(host); err == nil {
				ipPtr = &parsedIP
			}
		} else {
			if parsedIP, err := netip.ParseAddr(r.RemoteAddr); err == nil {
				ipPtr = &parsedIP
			}
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
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
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

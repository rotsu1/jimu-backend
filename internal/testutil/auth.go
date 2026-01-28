package testutil

import (
	"github.com/google/uuid"
	"github.com/rotsu1/jimu-backend/internal/auth"
)

// CreateTestToken generates a valid JWT access token for the given userID.
// The token is signed with TestJWTSecret and will be accepted by AuthMiddleware.
func CreateTestToken(userID uuid.UUID) string {
	accessToken, _, err := auth.GenerateTokenPair(userID.String(), TestJWTSecret)
	if err != nil {
		// In test context, panic is acceptable for setup failures
		panic("failed to create test token: " + err.Error())
	}
	return accessToken
}

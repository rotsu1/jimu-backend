package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// TestUser represents a user created for testing purposes.
type TestUser struct {
	ID       uuid.UUID
	Username string
}

// SeedUser creates a test user in the database and returns its info.
func (s *TestServer) SeedUser(t *testing.T, username string) *TestUser {
	t.Helper()

	var id uuid.UUID
	err := s.DB.QueryRow(
		context.Background(),
		"INSERT INTO profiles (username) VALUES ($1) RETURNING id",
		username,
	).Scan(&id)

	if err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	return &TestUser{
		ID:       id,
		Username: username,
	}
}

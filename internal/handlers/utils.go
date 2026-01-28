package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// GetIDFromRequest extracts a UUID from the request, preferring the last
// path segment if it's a valid UUID, falling back to query param "id".
func GetIDFromRequest(r *http.Request) (uuid.UUID, error) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		if id, err := uuid.Parse(lastPart); err == nil {
			return id, nil
		}
	}
	return uuid.Parse(r.URL.Query().Get("id"))
}

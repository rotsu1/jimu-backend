package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// GetIDFromRequest extracts a UUID from the request, preferring query param "id",
// falling back to the last path segment if it's a valid UUID.
func GetIDFromRequest(r *http.Request) (uuid.UUID, error) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			if _, err := uuid.Parse(lastPart); err == nil {
				idStr = lastPart
			}
		}
	}
	return uuid.Parse(idStr)
}

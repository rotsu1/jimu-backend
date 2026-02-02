package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var ErrMissingPathParam = errors.New("missing path param")

func pathParts(r *http.Request) []string {
	p := strings.Trim(r.URL.Path, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

// GetUUIDPathParam extracts a UUID from the request path at a specific segment
// index after trimming leading/trailing slashes.
//
// Example: "/workouts/{id}/likes" -> parts: ["workouts","{id}","likes"]
// GetUUIDPathParam(r, 1) returns the workout UUID.
func GetUUIDPathParam(r *http.Request, idx int) (uuid.UUID, error) {
	parts := pathParts(r)
	if idx < 0 || idx >= len(parts) {
		return uuid.Nil, ErrMissingPathParam
	}
	return uuid.Parse(parts[idx])
}

// GetIDFromRequest extracts a UUID from the request path's last segment.
// Per handler rule 2.1, resource identifiers MUST come from the path (not query).
func GetIDFromRequest(r *http.Request) (uuid.UUID, error) {
	parts := pathParts(r)
	if len(parts) == 0 {
		return uuid.Nil, ErrMissingPathParam
	}
	return uuid.Parse(parts[len(parts)-1])
}

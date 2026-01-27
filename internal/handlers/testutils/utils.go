package testutils

import (
	"context"
	"net/http"

	"github.com/rotsu1/jimu-backend/internal/middleware"
)

func InjectUserID(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

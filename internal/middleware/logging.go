package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "requestId"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = uuid.New().String() // Generate if not provided
		}

		// Add it to context
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)

		// Also set in response header (optional)
		w.Header().Set("X-Request-Id", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID Helper to get the requestId from context
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		return v.(string)
	}
	return ""
}

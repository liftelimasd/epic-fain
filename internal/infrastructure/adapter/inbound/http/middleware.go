package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

type contextKey string

const apiKeyActorKey contextKey = "actor"

// ActorFromContext extracts the authenticated actor (API key name) from context.
func ActorFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(apiKeyActorKey).(string); ok {
		return v
	}
	return "anonymous"
}

// APIKeyAuth is a middleware that validates the X-API-Key header.
// For v1 we use a simple in-memory map; production uses the api_keys table.
type APIKeyAuth struct {
	// keys maps SHA-256(key) → owner name
	keys map[string]string
}

func NewAPIKeyAuth(keys map[string]string) *APIKeyAuth {
	return &APIKeyAuth{keys: keys}
}

func (a *APIKeyAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}
		if key == "" {
			http.Error(w, `{"error":"missing API key"}`, http.StatusUnauthorized)
			return
		}

		hash := sha256Hash(key)
		owner, ok := a.keys[hash]
		if !ok {
			http.Error(w, `{"error":"invalid API key"}`, http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), apiKeyActorKey, owner)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(h[:])
}

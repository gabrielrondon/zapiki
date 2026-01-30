package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	APIKeyKey contextKey = "api_key"
)

// Auth provides API key authentication middleware
type Auth struct {
	apiKeyRepo *postgres.APIKeyRepository
}

// NewAuth creates a new auth middleware
func NewAuth(apiKeyRepo *postgres.APIKeyRepository) *Auth {
	return &Auth{
		apiKeyRepo: apiKeyRepo,
	}
}

// Authenticate validates the API key and adds user context
func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Try Authorization header
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			http.Error(w, `{"error":"Missing API key"}`, http.StatusUnauthorized)
			return
		}

		// Validate API key
		key, err := a.apiKeyRepo.GetByKey(r.Context(), apiKey)
		if err != nil {
			http.Error(w, `{"error":"Invalid API key"}`, http.StatusUnauthorized)
			return
		}

		// Update last used timestamp (async, don't wait)
		go a.apiKeyRepo.UpdateLastUsed(context.Background(), apiKey)

		// Add user ID to context
		ctx := context.WithValue(r.Context(), UserIDKey, key.UserID)
		ctx = context.WithValue(ctx, APIKeyKey, key)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, http.ErrNoCookie
	}
	return userID, nil
}

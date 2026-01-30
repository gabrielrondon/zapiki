package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/storage/redis"
)

// RateLimit provides rate limiting middleware
type RateLimit struct {
	limiter *redis.RateLimiter
	window  time.Duration
}

// NewRateLimit creates a new rate limit middleware
func NewRateLimit(limiter *redis.RateLimiter, window time.Duration) *RateLimit {
	return &RateLimit{
		limiter: limiter,
		window:  window,
	}
}

// Limit applies rate limiting based on API key
func (rl *RateLimit) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from context
		apiKey, ok := r.Context().Value(APIKeyKey).(*models.APIKey)
		if !ok {
			http.Error(w, `{"error":"API key not found in context"}`, http.StatusInternalServerError)
			return
		}

		// Create rate limit key
		key := fmt.Sprintf("ratelimit:%s", apiKey.Key)

		// Check rate limit
		allowed, err := rl.limiter.Allow(r.Context(), key, apiKey.RateLimit, rl.window)
		if err != nil {
			http.Error(w, `{"error":"Rate limit check failed"}`, http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", apiKey.RateLimit))
			w.Header().Set("X-RateLimit-Window", rl.window.String())
			http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

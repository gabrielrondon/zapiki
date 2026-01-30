package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gabrielrondon/zapiki/internal/config"
	"github.com/redis/go-redis/v9"
)

// Store wraps a Redis client
type Store struct {
	client *redis.Client
}

// New creates a new Redis store
func New(cfg *config.RedisConfig) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Store{client: client}, nil
}

// Close closes the Redis connection
func (s *Store) Close() error {
	return s.client.Close()
}

// Client returns the underlying Redis client
func (s *Store) Client() *redis.Client {
	return s.client
}

// Health checks if Redis is healthy
func (s *Store) Health(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	store *Store
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(store *Store) *RateLimiter {
	return &RateLimiter{store: store}
}

// Allow checks if a request is allowed based on rate limit
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-window).UnixMilli()

	pipe := rl.store.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiration
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	count := countCmd.Val()
	return count < int64(limit), nil
}

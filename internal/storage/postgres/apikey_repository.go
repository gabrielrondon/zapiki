package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
)

// APIKeyRepository handles API key database operations
type APIKeyRepository struct {
	store *Store
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(store *Store) *APIKeyRepository {
	return &APIKeyRepository{store: store}
}

// GetByKey retrieves an API key by its key value
func (r *APIKeyRepository) GetByKey(ctx context.Context, key string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, key, name, rate_limit, is_active, last_used_at, created_at, expires_at
		FROM api_keys
		WHERE key = $1 AND is_active = true
	`

	var apiKey models.APIKey
	err := r.store.pool.QueryRow(ctx, query, key).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.Name,
		&apiKey.RateLimit, &apiKey.IsActive, &apiKey.LastUsedAt,
		&apiKey.CreatedAt, &apiKey.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key expired")
	}

	return &apiKey, nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, key string) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE key = $1`

	_, err := r.store.pool.Exec(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to update last used time: %w", err)
	}

	return nil
}

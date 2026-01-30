package postgres

import (
	"context"
	"fmt"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/google/uuid"
)

// CircuitRepository handles circuit database operations
type CircuitRepository struct {
	store *Store
}

// NewCircuitRepository creates a new circuit repository
func NewCircuitRepository(store *Store) *CircuitRepository {
	return &CircuitRepository{store: store}
}

// Create creates a new circuit record
func (r *CircuitRepository) Create(ctx context.Context, circuit *models.Circuit) error {
	query := `
		INSERT INTO circuits (
			id, user_id, name, description, proof_system,
			circuit_definition, proving_key_url, verification_key_url,
			is_public, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
		)
	`

	_, err := r.store.pool.Exec(ctx, query,
		circuit.ID, circuit.UserID, circuit.Name, circuit.Description,
		circuit.ProofSystem, circuit.CircuitDefinition,
		circuit.ProvingKeyURL, circuit.VerificationKeyURL,
		circuit.IsPublic,
	)

	if err != nil {
		return fmt.Errorf("failed to create circuit: %w", err)
	}

	return nil
}

// GetByID retrieves a circuit by ID
func (r *CircuitRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Circuit, error) {
	query := `
		SELECT id, user_id, name, description, proof_system,
			   circuit_definition, proving_key_url, verification_key_url,
			   is_public, created_at, updated_at
		FROM circuits
		WHERE id = $1
	`

	var circuit models.Circuit
	err := r.store.pool.QueryRow(ctx, query, id).Scan(
		&circuit.ID, &circuit.UserID, &circuit.Name, &circuit.Description,
		&circuit.ProofSystem, &circuit.CircuitDefinition,
		&circuit.ProvingKeyURL, &circuit.VerificationKeyURL,
		&circuit.IsPublic, &circuit.CreatedAt, &circuit.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get circuit: %w", err)
	}

	return &circuit, nil
}

// ListByUser retrieves circuits for a user
func (r *CircuitRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Circuit, error) {
	query := `
		SELECT id, user_id, name, description, proof_system,
			   circuit_definition, proving_key_url, verification_key_url,
			   is_public, created_at, updated_at
		FROM circuits
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.store.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list circuits: %w", err)
	}
	defer rows.Close()

	var circuits []*models.Circuit
	for rows.Next() {
		var circuit models.Circuit
		err := rows.Scan(
			&circuit.ID, &circuit.UserID, &circuit.Name, &circuit.Description,
			&circuit.ProofSystem, &circuit.CircuitDefinition,
			&circuit.ProvingKeyURL, &circuit.VerificationKeyURL,
			&circuit.IsPublic, &circuit.CreatedAt, &circuit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan circuit: %w", err)
		}
		circuits = append(circuits, &circuit)
	}

	return circuits, nil
}

// ListByUserWithPublic retrieves user's circuits plus public circuits
func (r *CircuitRepository) ListByUserWithPublic(ctx context.Context, userID uuid.UUID) ([]*models.Circuit, error) {
	query := `
		SELECT id, user_id, name, description, proof_system,
			   circuit_definition, proving_key_url, verification_key_url,
			   is_public, created_at, updated_at
		FROM circuits
		WHERE user_id = $1 OR is_public = true
		ORDER BY created_at DESC
	`

	rows, err := r.store.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list circuits: %w", err)
	}
	defer rows.Close()

	var circuits []*models.Circuit
	for rows.Next() {
		var circuit models.Circuit
		err := rows.Scan(
			&circuit.ID, &circuit.UserID, &circuit.Name, &circuit.Description,
			&circuit.ProofSystem, &circuit.CircuitDefinition,
			&circuit.ProvingKeyURL, &circuit.VerificationKeyURL,
			&circuit.IsPublic, &circuit.CreatedAt, &circuit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan circuit: %w", err)
		}
		circuits = append(circuits, &circuit)
	}

	return circuits, nil
}

// Delete deletes a circuit by ID
func (r *CircuitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM circuits WHERE id = $1`

	result, err := r.store.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete circuit: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("circuit not found")
	}

	return nil
}

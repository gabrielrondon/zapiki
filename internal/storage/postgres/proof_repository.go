package postgres

import (
	"context"
	"fmt"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/google/uuid"
)

// ProofRepository handles proof database operations
type ProofRepository struct {
	store *Store
}

// NewProofRepository creates a new proof repository
func NewProofRepository(store *Store) *ProofRepository {
	return &ProofRepository{store: store}
}

// Create creates a new proof record
func (r *ProofRepository) Create(ctx context.Context, proof *models.Proof) error {
	query := `
		INSERT INTO proofs (
			id, user_id, circuit_id, template_id, proof_system, status,
			input_data, proof_data, public_inputs, proof_url, error_message,
			generation_time_ms, created_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.store.pool.Exec(ctx, query,
		proof.ID, proof.UserID, proof.CircuitID, proof.TemplateID,
		proof.ProofSystem, proof.Status, proof.InputData, proof.ProofData,
		proof.PublicInputs, proof.ProofURL, proof.ErrorMessage,
		proof.GenerationTimeMs, proof.CreatedAt, proof.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create proof: %w", err)
	}

	return nil
}

// GetByID retrieves a proof by ID
func (r *ProofRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Proof, error) {
	query := `
		SELECT id, user_id, circuit_id, template_id, proof_system, status,
			   input_data, proof_data, public_inputs, proof_url, error_message,
			   generation_time_ms, created_at, completed_at
		FROM proofs
		WHERE id = $1
	`

	var proof models.Proof
	err := r.store.pool.QueryRow(ctx, query, id).Scan(
		&proof.ID, &proof.UserID, &proof.CircuitID, &proof.TemplateID,
		&proof.ProofSystem, &proof.Status, &proof.InputData, &proof.ProofData,
		&proof.PublicInputs, &proof.ProofURL, &proof.ErrorMessage,
		&proof.GenerationTimeMs, &proof.CreatedAt, &proof.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}

	return &proof, nil
}

// Update updates a proof record
func (r *ProofRepository) Update(ctx context.Context, proof *models.Proof) error {
	query := `
		UPDATE proofs
		SET status = $1, proof_data = $2, public_inputs = $3, proof_url = $4,
		    error_message = $5, generation_time_ms = $6, completed_at = $7
		WHERE id = $8
	`

	result, err := r.store.pool.Exec(ctx, query,
		proof.Status, proof.ProofData, proof.PublicInputs, proof.ProofURL,
		proof.ErrorMessage, proof.GenerationTimeMs, proof.CompletedAt, proof.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update proof: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("proof not found")
	}

	return nil
}

// ListByUser retrieves proofs for a user
func (r *ProofRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Proof, error) {
	query := `
		SELECT id, user_id, circuit_id, template_id, proof_system, status,
			   input_data, proof_data, public_inputs, proof_url, error_message,
			   generation_time_ms, created_at, completed_at
		FROM proofs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.store.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list proofs: %w", err)
	}
	defer rows.Close()

	var proofs []*models.Proof
	for rows.Next() {
		var proof models.Proof
		err := rows.Scan(
			&proof.ID, &proof.UserID, &proof.CircuitID, &proof.TemplateID,
			&proof.ProofSystem, &proof.Status, &proof.InputData, &proof.ProofData,
			&proof.PublicInputs, &proof.ProofURL, &proof.ErrorMessage,
			&proof.GenerationTimeMs, &proof.CreatedAt, &proof.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proof: %w", err)
		}
		proofs = append(proofs, &proof)
	}

	return proofs, nil
}

// Delete deletes a proof by ID
func (r *ProofRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM proofs WHERE id = $1`

	result, err := r.store.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete proof: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("proof not found")
	}

	return nil
}

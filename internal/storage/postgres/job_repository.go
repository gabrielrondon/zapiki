package postgres

import (
	"context"
	"fmt"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/google/uuid"
)

// JobRepository handles job database operations
type JobRepository struct {
	store *Store
}

// NewJobRepository creates a new job repository
func NewJobRepository(store *Store) *JobRepository {
	return &JobRepository{store: store}
}

// Create creates a new job record
func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (
			id, user_id, proof_id, status, priority, retry_count,
			max_retries, error_message, created_at, started_at, completed_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.store.pool.Exec(ctx, query,
		job.ID, job.UserID, job.ProofID, job.Status, job.Priority,
		job.RetryCount, job.MaxRetries, job.ErrorMessage,
		job.CreatedAt, job.StartedAt, job.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetByID retrieves a job by ID
func (r *JobRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	query := `
		SELECT id, user_id, proof_id, status, priority, retry_count,
			   max_retries, error_message, created_at, started_at, completed_at
		FROM jobs
		WHERE id = $1
	`

	var job models.Job
	err := r.store.pool.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.UserID, &job.ProofID, &job.Status, &job.Priority,
		&job.RetryCount, &job.MaxRetries, &job.ErrorMessage,
		&job.CreatedAt, &job.StartedAt, &job.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// GetByProofID retrieves a job by proof ID
func (r *JobRepository) GetByProofID(ctx context.Context, proofID uuid.UUID) (*models.Job, error) {
	query := `
		SELECT id, user_id, proof_id, status, priority, retry_count,
			   max_retries, error_message, created_at, started_at, completed_at
		FROM jobs
		WHERE proof_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var job models.Job
	err := r.store.pool.QueryRow(ctx, query, proofID).Scan(
		&job.ID, &job.UserID, &job.ProofID, &job.Status, &job.Priority,
		&job.RetryCount, &job.MaxRetries, &job.ErrorMessage,
		&job.CreatedAt, &job.StartedAt, &job.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// Update updates a job record
func (r *JobRepository) Update(ctx context.Context, job *models.Job) error {
	query := `
		UPDATE jobs
		SET status = $1, retry_count = $2, error_message = $3,
		    started_at = $4, completed_at = $5
		WHERE id = $6
	`

	result, err := r.store.pool.Exec(ctx, query,
		job.Status, job.RetryCount, job.ErrorMessage,
		job.StartedAt, job.CompletedAt, job.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

// ListByUser retrieves jobs for a user
func (r *JobRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Job, error) {
	query := `
		SELECT id, user_id, proof_id, status, priority, retry_count,
			   max_retries, error_message, created_at, started_at, completed_at
		FROM jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.store.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		var job models.Job
		err := rows.Scan(
			&job.ID, &job.UserID, &job.ProofID, &job.Status, &job.Priority,
			&job.RetryCount, &job.MaxRetries, &job.ErrorMessage,
			&job.CreatedAt, &job.StartedAt, &job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

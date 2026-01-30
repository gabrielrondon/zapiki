package postgres

import (
	"context"
	"fmt"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/google/uuid"
)

// TemplateRepository handles template database operations
type TemplateRepository struct {
	store *Store
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(store *Store) *TemplateRepository {
	return &TemplateRepository{store: store}
}

// Create creates a new template record
func (r *TemplateRepository) Create(ctx context.Context, template *models.Template) error {
	query := `
		INSERT INTO templates (
			id, name, description, category, proof_system,
			circuit_id, input_schema, example_inputs,
			documentation, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
		)
	`

	_, err := r.store.pool.Exec(ctx, query,
		template.ID, template.Name, template.Description,
		template.Category, template.ProofSystem, template.CircuitID,
		template.InputSchema, template.ExampleInputs,
		template.Documentation, template.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

// GetByID retrieves a template by ID
func (r *TemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	query := `
		SELECT id, name, description, category, proof_system,
			   circuit_id, input_schema, example_inputs,
			   documentation, is_active, created_at, updated_at
		FROM templates
		WHERE id = $1
	`

	var template models.Template
	err := r.store.pool.QueryRow(ctx, query, id).Scan(
		&template.ID, &template.Name, &template.Description,
		&template.Category, &template.ProofSystem, &template.CircuitID,
		&template.InputSchema, &template.ExampleInputs,
		&template.Documentation, &template.IsActive,
		&template.CreatedAt, &template.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &template, nil
}

// ListActive retrieves all active templates
func (r *TemplateRepository) ListActive(ctx context.Context) ([]*models.Template, error) {
	query := `
		SELECT id, name, description, category, proof_system,
			   circuit_id, input_schema, example_inputs,
			   documentation, is_active, created_at, updated_at
		FROM templates
		WHERE is_active = true
		ORDER BY category, name
	`

	rows, err := r.store.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.Template
	for rows.Next() {
		var template models.Template
		err := rows.Scan(
			&template.ID, &template.Name, &template.Description,
			&template.Category, &template.ProofSystem, &template.CircuitID,
			&template.InputSchema, &template.ExampleInputs,
			&template.Documentation, &template.IsActive,
			&template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, &template)
	}

	return templates, nil
}

// ListByCategory retrieves templates by category
func (r *TemplateRepository) ListByCategory(ctx context.Context, category string) ([]*models.Template, error) {
	query := `
		SELECT id, name, description, category, proof_system,
			   circuit_id, input_schema, example_inputs,
			   documentation, is_active, created_at, updated_at
		FROM templates
		WHERE category = $1 AND is_active = true
		ORDER BY name
	`

	rows, err := r.store.pool.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.Template
	for rows.Next() {
		var template models.Template
		err := rows.Scan(
			&template.ID, &template.Name, &template.Description,
			&template.Category, &template.ProofSystem, &template.CircuitID,
			&template.InputSchema, &template.ExampleInputs,
			&template.Documentation, &template.IsActive,
			&template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, &template)
	}

	return templates, nil
}

// GetCategories returns all distinct categories
func (r *TemplateRepository) GetCategories(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT category
		FROM templates
		WHERE is_active = true
		ORDER BY category
	`

	rows, err := r.store.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// Update updates a template
func (r *TemplateRepository) Update(ctx context.Context, template *models.Template) error {
	query := `
		UPDATE templates
		SET name = $1, description = $2, category = $3,
		    input_schema = $4, example_inputs = $5,
		    documentation = $6, is_active = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := r.store.pool.Exec(ctx, query,
		template.Name, template.Description, template.Category,
		template.InputSchema, template.ExampleInputs,
		template.Documentation, template.IsActive, template.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// Delete deletes a template (soft delete by setting is_active = false)
func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE templates SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.store.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

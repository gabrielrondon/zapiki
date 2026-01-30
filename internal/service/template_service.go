package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/google/uuid"
)

// TemplateService handles template management
type TemplateService struct {
	templateRepo *postgres.TemplateRepository
	circuitRepo  *postgres.CircuitRepository
	proofService *ProofService
}

// NewTemplateService creates a new template service
func NewTemplateService(
	templateRepo *postgres.TemplateRepository,
	circuitRepo *postgres.CircuitRepository,
	proofService *ProofService,
) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		circuitRepo:  circuitRepo,
		proofService: proofService,
	}
}

// GenerateFromTemplateRequest represents a request to generate proof from template
type GenerateFromTemplateRequest struct {
	UserID   uuid.UUID              `json:"user_id"`
	Inputs   map[string]interface{} `json:"inputs"`
	Options  *models.ProofOptions   `json:"options,omitempty"`
}

// GenerateFromTemplate generates a proof using a template
func (s *TemplateService) GenerateFromTemplate(
	ctx context.Context,
	templateID uuid.UUID,
	req *GenerateFromTemplateRequest,
) (*GenerateProofResponse, error) {
	// Get template
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("template is not active")
	}

	// Validate inputs against schema
	if err := s.validateInputs(template, req.Inputs); err != nil {
		return nil, fmt.Errorf("invalid inputs: %w", err)
	}

	// Get circuit
	circuit, err := s.circuitRepo.GetByID(ctx, template.CircuitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit: %w", err)
	}

	// Convert inputs to data format
	inputJSON, err := json.Marshal(req.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs: %w", err)
	}

	// Create proof request
	proofReq := &GenerateProofRequest{
		UserID:      req.UserID,
		ProofSystem: template.ProofSystem,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: inputJSON,
		},
		Options: req.Options,
	}

	// Ensure circuit ID is set
	if proofReq.Options == nil {
		proofReq.Options = &models.ProofOptions{}
	}
	proofReq.Options.CircuitID = &circuit.ID
	proofReq.Options.TemplateID = &template.ID

	// Generate proof
	return s.proofService.Generate(ctx, proofReq)
}

// List lists all active templates
func (s *TemplateService) List(ctx context.Context, category string) ([]*models.Template, error) {
	if category != "" {
		return s.templateRepo.ListByCategory(ctx, category)
	}
	return s.templateRepo.ListActive(ctx)
}

// Get retrieves a template by ID
func (s *TemplateService) Get(ctx context.Context, templateID uuid.UUID) (*models.Template, error) {
	return s.templateRepo.GetByID(ctx, templateID)
}

// validateInputs validates inputs against template schema
func (s *TemplateService) validateInputs(template *models.Template, inputs map[string]interface{}) error {
	// Parse schema
	var schema map[string]interface{}
	if err := json.Unmarshal(template.InputSchema, &schema); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Get required fields
	required, ok := schema["required"].([]interface{})
	if !ok {
		return nil // No required fields
	}

	// Check all required fields are present
	for _, field := range required {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}

		if _, exists := inputs[fieldName]; !exists {
			return fmt.Errorf("missing required field: %s", fieldName)
		}
	}

	return nil
}

// GetCategories returns all template categories
func (s *TemplateService) GetCategories(ctx context.Context) ([]string, error) {
	return s.templateRepo.GetCategories(ctx)
}

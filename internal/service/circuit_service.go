package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/google/uuid"
)

// CircuitService handles circuit management logic
type CircuitService struct {
	factory     *prover.Factory
	circuitRepo *postgres.CircuitRepository
}

// NewCircuitService creates a new circuit service
func NewCircuitService(factory *prover.Factory, circuitRepo *postgres.CircuitRepository) *CircuitService {
	return &CircuitService{
		factory:     factory,
		circuitRepo: circuitRepo,
	}
}

// CreateCircuitRequest represents a request to create a circuit
type CreateCircuitRequest struct {
	UserID            uuid.UUID              `json:"user_id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	ProofSystem       models.ProofSystemType `json:"proof_system"`
	CircuitDefinition json.RawMessage        `json:"circuit_definition"`
	IsPublic          bool                   `json:"is_public"`
}

// CreateCircuitResponse represents the response from circuit creation
type CreateCircuitResponse struct {
	Circuit         *models.Circuit `json:"circuit"`
	SetupRequired   bool            `json:"setup_required"`
	SetupInProgress bool            `json:"setup_in_progress"`
}

// Create creates a new circuit and optionally runs setup
func (s *CircuitService) Create(ctx context.Context, req *CreateCircuitRequest) (*CreateCircuitResponse, error) {
	// Validate proof system
	system, err := s.factory.Get(req.ProofSystem)
	if err != nil {
		return nil, fmt.Errorf("unsupported proof system: %w", err)
	}

	// Create circuit record
	circuit := &models.Circuit{
		ID:                uuid.New(),
		UserID:            req.UserID,
		Name:              req.Name,
		Description:       req.Description,
		ProofSystem:       req.ProofSystem,
		CircuitDefinition: req.CircuitDefinition,
		IsPublic:          req.IsPublic,
	}

	// Check if setup is required
	caps := system.Capabilities()
	setupRequired := caps.SupportsSetup

	if setupRequired {
		// Run setup
		_, err := system.Setup(ctx, circuit)
		if err != nil {
			return nil, fmt.Errorf("failed to run setup: %w", err)
		}

		// Store keys (in production, upload to S3)
		// For now, store serialized keys in database URLs
		circuit.ProvingKeyURL = fmt.Sprintf("db:%s:pk", circuit.ID)
		circuit.VerificationKeyURL = fmt.Sprintf("db:%s:vk", circuit.ID)

		// TODO: Upload to S3 in production
		// circuit.ProvingKeyURL = s.uploadToS3(setupResult.ProvingKey)
		// circuit.VerificationKeyURL = s.uploadToS3(setupResult.VerificationKey)
	}

	// Save circuit to database
	if err := s.circuitRepo.Create(ctx, circuit); err != nil {
		return nil, fmt.Errorf("failed to create circuit: %w", err)
	}

	return &CreateCircuitResponse{
		Circuit:         circuit,
		SetupRequired:   setupRequired,
		SetupInProgress: false,
	}, nil
}

// Get retrieves a circuit by ID
func (s *CircuitService) Get(ctx context.Context, circuitID uuid.UUID, userID uuid.UUID) (*models.Circuit, error) {
	circuit, err := s.circuitRepo.GetByID(ctx, circuitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit: %w", err)
	}

	// Check access: owner or public circuit
	if circuit.UserID != userID && !circuit.IsPublic {
		return nil, fmt.Errorf("unauthorized: circuit is private")
	}

	return circuit, nil
}

// List lists circuits for a user
func (s *CircuitService) List(ctx context.Context, userID uuid.UUID, includePublic bool) ([]*models.Circuit, error) {
	if includePublic {
		return s.circuitRepo.ListByUserWithPublic(ctx, userID)
	}
	return s.circuitRepo.ListByUser(ctx, userID)
}

// Delete deletes a circuit
func (s *CircuitService) Delete(ctx context.Context, circuitID uuid.UUID, userID uuid.UUID) error {
	// Verify ownership
	circuit, err := s.circuitRepo.GetByID(ctx, circuitID)
	if err != nil {
		return fmt.Errorf("failed to get circuit: %w", err)
	}

	if circuit.UserID != userID {
		return fmt.Errorf("unauthorized: circuit belongs to different user")
	}

	return s.circuitRepo.Delete(ctx, circuitID)
}

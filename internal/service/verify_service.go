package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/google/uuid"
)

// VerifyService handles proof verification logic
type VerifyService struct {
	factory *prover.Factory
}

// NewVerifyService creates a new verify service
func NewVerifyService(factory *prover.Factory) *VerifyService {
	return &VerifyService{
		factory: factory,
	}
}

// VerifyRequest represents a verification request
type VerifyRequest struct {
	ProofSystem     models.ProofSystemType `json:"proof_system"`
	Proof           json.RawMessage        `json:"proof"`
	VerificationKey json.RawMessage        `json:"verification_key"`
	PublicInputs    json.RawMessage        `json:"public_inputs,omitempty"`
}

// VerifyResponse represents a verification response
type VerifyResponse struct {
	Valid        bool   `json:"valid"`
	ErrorMessage string `json:"error_message,omitempty"`
	VerifiedAt   time.Time `json:"verified_at"`
}

// Verify verifies a proof
func (s *VerifyService) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	// Get the proof system
	system, err := s.factory.Get(req.ProofSystem)
	if err != nil {
		return nil, fmt.Errorf("unsupported proof system: %w", err)
	}

	// Verify the proof
	proverReq := &prover.VerifyRequest{
		Proof:           req.Proof,
		VerificationKey: req.VerificationKey,
		PublicInputs:    req.PublicInputs,
	}

	proverResp, err := system.Verify(ctx, proverReq)
	if err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	return &VerifyResponse{
		Valid:        proverResp.Valid,
		ErrorMessage: proverResp.ErrorMessage,
		VerifiedAt:   time.Now(),
	}, nil
}

// VerifyProofByID verifies a proof by its ID
func (s *VerifyService) VerifyProofByID(ctx context.Context, proofID uuid.UUID) (*VerifyResponse, error) {
	// This would require fetching the proof from the database
	// For now, return not implemented
	return nil, fmt.Errorf("not implemented: use direct verification endpoint")
}

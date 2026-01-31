package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/gabrielrondon/zapiki/internal/queue"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/google/uuid"
)

// ProofService handles proof generation logic
type ProofService struct {
	factory    *prover.Factory
	proofRepo  *postgres.ProofRepository
	jobRepo    *postgres.JobRepository
	queueClient interface {
		EnqueueProofGeneration(ctx context.Context, payload interface{}, priority int) error
	}
}

// NewProofService creates a new proof service
func NewProofService(factory *prover.Factory, proofRepo *postgres.ProofRepository, jobRepo *postgres.JobRepository, queueClient interface {
	EnqueueProofGeneration(ctx context.Context, payload interface{}, priority int) error
}) *ProofService {
	return &ProofService{
		factory:     factory,
		proofRepo:   proofRepo,
		jobRepo:     jobRepo,
		queueClient: queueClient,
	}
}

// GenerateProofRequest represents a request to generate a proof
type GenerateProofRequest struct {
	UserID       uuid.UUID              `json:"user_id"`
	ProofSystem  models.ProofSystemType `json:"proof_system"`
	Data         *models.InputData      `json:"data"`
	PublicInputs json.RawMessage        `json:"public_inputs,omitempty"`
	Options      *models.ProofOptions   `json:"options,omitempty"`
}

// GenerateProofResponse represents the response from proof generation
type GenerateProofResponse struct {
	ProofID          uuid.UUID            `json:"proof_id"`
	Status           models.ProofStatus   `json:"status"`
	Proof            json.RawMessage      `json:"proof,omitempty"`
	VerificationKey  json.RawMessage      `json:"verification_key,omitempty"`
	GenerationTimeMs int64                `json:"generation_time_ms,omitempty"`
	Message          string               `json:"message,omitempty"`
}

// Generate generates a proof
func (s *ProofService) Generate(ctx context.Context, req *GenerateProofRequest) (*GenerateProofResponse, error) {
	// Get the proof system
	system, err := s.factory.Get(req.ProofSystem)
	if err != nil {
		return nil, fmt.Errorf("unsupported proof system: %w", err)
	}

	// Create proof record
	proofID := uuid.New()
	proof := &models.Proof{
		ID:           proofID,
		UserID:       req.UserID,
		ProofSystem:  req.ProofSystem,
		Status:       models.ProofStatusPending,
		InputData:    nil, // Don't store sensitive data by default
		PublicInputs: req.PublicInputs,
		CreatedAt:    time.Now(),
	}

	if req.Options != nil {
		proof.CircuitID = req.Options.CircuitID
		proof.TemplateID = req.Options.TemplateID
	}

	// Check if this should be async
	caps := system.Capabilities()
	isAsync := caps.AsyncOnly || (req.Options != nil && req.Options.Async)

	if isAsync {
		// For async processing, create the proof record and enqueue job
		proof.Status = models.ProofStatusPending
		if err := s.proofRepo.Create(ctx, proof); err != nil {
			return nil, fmt.Errorf("failed to create proof record: %w", err)
		}

		// Create job record
		job := &models.Job{
			ID:         uuid.New(),
			UserID:     req.UserID,
			ProofID:    proofID,
			Status:     models.ProofStatusPending,
			Priority:   0,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		}

		if err := s.jobRepo.Create(ctx, job); err != nil {
			return nil, fmt.Errorf("failed to create job record: %w", err)
		}

		// Enqueue job if queue client is available
		if s.queueClient != nil {
			queuePayload := &queue.ProofGenerationPayload{
				ProofID:      proofID,
				UserID:       req.UserID,
				ProofSystem:  req.ProofSystem,
				Data:         req.Data,
				PublicInputs: req.PublicInputs,
			}

			if req.Options != nil {
				queuePayload.CircuitID = req.Options.CircuitID
				queuePayload.TemplateID = req.Options.TemplateID
				// Pass additional options
				queuePayload.Options = map[string]interface{}{
					"async": req.Options.Async,
				}
			}

			if err := s.queueClient.EnqueueProofGeneration(ctx, queuePayload, job.Priority); err != nil {
				return nil, fmt.Errorf("failed to enqueue job: %w", err)
			}
		}

		return &GenerateProofResponse{
			ProofID: proofID,
			Status:  models.ProofStatusPending,
			Message: "Proof generation started. Poll /api/v1/proofs/" + proofID.String() + " for status.",
		}, nil
	}

	// Synchronous proof generation
	proof.Status = models.ProofStatusProcessing
	if err := s.proofRepo.Create(ctx, proof); err != nil {
		return nil, fmt.Errorf("failed to create proof record: %w", err)
	}

	// Generate the proof
	startTime := time.Now()
	proverReq := &prover.ProofRequest{
		Data:         req.Data,
		PublicInputs: req.PublicInputs,
		Options:      make(map[string]interface{}),
	}

	// Pass circuit/template information to prover
	if req.Options != nil {
		if req.Options.CircuitID != nil {
			proverReq.Options["circuit_id"] = req.Options.CircuitID
		}
		if req.Options.TemplateID != nil {
			proverReq.Options["template_id"] = req.Options.TemplateID
		}
		proverReq.Options["async"] = req.Options.Async
	}

	proverResp, err := system.Generate(ctx, proverReq)
	if err != nil {
		// Update proof with error
		proof.Status = models.ProofStatusFailed
		proof.ErrorMessage = err.Error()
		now := time.Now()
		proof.CompletedAt = &now
		_ = s.proofRepo.Update(ctx, proof)

		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	// Update proof with result
	proof.Status = models.ProofStatusCompleted
	proof.ProofData = proverResp.Proof
	proof.PublicInputs = proverResp.PublicInputs
	proof.GenerationTimeMs = time.Since(startTime).Milliseconds()
	now := time.Now()
	proof.CompletedAt = &now

	if err := s.proofRepo.Update(ctx, proof); err != nil {
		return nil, fmt.Errorf("failed to update proof record: %w", err)
	}

	return &GenerateProofResponse{
		ProofID:          proofID,
		Status:           models.ProofStatusCompleted,
		Proof:            proverResp.Proof,
		VerificationKey:  proverResp.VerificationKey,
		GenerationTimeMs: proof.GenerationTimeMs,
	}, nil
}

// GetProof retrieves a proof by ID
func (s *ProofService) GetProof(ctx context.Context, proofID uuid.UUID, userID uuid.UUID) (*models.Proof, error) {
	proof, err := s.proofRepo.GetByID(ctx, proofID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}

	// Verify user ownership
	if proof.UserID != userID {
		return nil, fmt.Errorf("unauthorized: proof belongs to different user")
	}

	return proof, nil
}

// ListProofs lists proofs for a user
func (s *ProofService) ListProofs(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Proof, error) {
	return s.proofRepo.ListByUser(ctx, userID, limit, offset)
}

// DeleteProof deletes a proof
func (s *ProofService) DeleteProof(ctx context.Context, proofID uuid.UUID, userID uuid.UUID) error {
	// Verify ownership first
	proof, err := s.proofRepo.GetByID(ctx, proofID)
	if err != nil {
		return fmt.Errorf("failed to get proof: %w", err)
	}

	if proof.UserID != userID {
		return fmt.Errorf("unauthorized: proof belongs to different user")
	}

	return s.proofRepo.Delete(ctx, proofID)
}

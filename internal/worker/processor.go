package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/gabrielrondon/zapiki/internal/queue"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/hibiken/asynq"
)

// Processor handles background job processing
type Processor struct {
	factory   *prover.Factory
	proofRepo *postgres.ProofRepository
	jobRepo   *postgres.JobRepository
}

// NewProcessor creates a new job processor
func NewProcessor(factory *prover.Factory, proofRepo *postgres.ProofRepository, jobRepo *postgres.JobRepository) *Processor {
	return &Processor{
		factory:   factory,
		proofRepo: proofRepo,
		jobRepo:   jobRepo,
	}
}

// HandleProofGeneration processes proof generation jobs
func (p *Processor) HandleProofGeneration(ctx context.Context, task *asynq.Task) error {
	// Parse payload
	var payload queue.ProofGenerationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	fmt.Printf("Processing proof generation: %s (system: %s)\n", payload.ProofID, payload.ProofSystem)

	// Update proof status to processing
	proof, err := p.proofRepo.GetByID(ctx, payload.ProofID)
	if err != nil {
		return fmt.Errorf("failed to get proof: %w", err)
	}

	proof.Status = models.ProofStatusProcessing
	if err := p.proofRepo.Update(ctx, proof); err != nil {
		return fmt.Errorf("failed to update proof status: %w", err)
	}

	// Update job status to processing
	job, err := p.jobRepo.GetByProofID(ctx, payload.ProofID)
	if err == nil && job != nil {
		now := time.Now()
		job.Status = models.ProofStatusProcessing
		job.StartedAt = &now
		_ = p.jobRepo.Update(ctx, job)
	}

	// Get proof system
	system, err := p.factory.Get(payload.ProofSystem)
	if err != nil {
		return p.handleError(ctx, proof, job, fmt.Errorf("unsupported proof system: %w", err))
	}

	// Generate proof
	startTime := time.Now()
	proverReq := &prover.ProofRequest{
		Data:         payload.Data,
		PublicInputs: payload.PublicInputs,
		Options:      payload.Options,
	}

	// Add circuit/template info to options if available
	if proverReq.Options == nil {
		proverReq.Options = make(map[string]interface{})
	}
	if payload.CircuitID != nil {
		proverReq.Options["circuit_id"] = payload.CircuitID
	}
	if payload.TemplateID != nil {
		proverReq.Options["template_id"] = payload.TemplateID
	}

	proverResp, err := system.Generate(ctx, proverReq)
	if err != nil {
		return p.handleError(ctx, proof, job, fmt.Errorf("failed to generate proof: %w", err))
	}

	// Update proof with result
	proof.Status = models.ProofStatusCompleted
	proof.ProofData = proverResp.Proof
	proof.PublicInputs = proverResp.PublicInputs
	proof.GenerationTimeMs = time.Since(startTime).Milliseconds()
	now := time.Now()
	proof.CompletedAt = &now

	if err := p.proofRepo.Update(ctx, proof); err != nil {
		return fmt.Errorf("failed to update proof: %w", err)
	}

	// Update job status
	if job != nil {
		job.Status = models.ProofStatusCompleted
		job.CompletedAt = &now
		_ = p.jobRepo.Update(ctx, job)
	}

	fmt.Printf("Proof generation completed: %s (took %dms)\n", payload.ProofID, proof.GenerationTimeMs)
	return nil
}

// handleError handles job processing errors
func (p *Processor) handleError(ctx context.Context, proof *models.Proof, job *models.Job, err error) error {
	// Update proof status
	proof.Status = models.ProofStatusFailed
	proof.ErrorMessage = err.Error()
	now := time.Now()
	proof.CompletedAt = &now
	_ = p.proofRepo.Update(ctx, proof)

	// Update job status
	if job != nil {
		job.Status = models.ProofStatusFailed
		job.ErrorMessage = err.Error()
		job.CompletedAt = &now
		_ = p.jobRepo.Update(ctx, job)
	}

	return err
}

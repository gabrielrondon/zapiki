package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	// Task types
	TypeProofGeneration = "proof:generate"
)

// Client wraps an asynq client for enqueueing jobs
type Client struct {
	client *asynq.Client
}

// NewClient creates a new queue client
func NewClient(redisAddr, password string) *Client {
	opt := asynq.RedisClientOpt{
		Addr: redisAddr,
	}
	if password != "" {
		opt.Password = password
	}

	client := asynq.NewClient(opt)

	return &Client{client: client}
}

// Close closes the queue client
func (c *Client) Close() error {
	return c.client.Close()
}

// ProofGenerationPayload represents the payload for proof generation jobs
type ProofGenerationPayload struct {
	ProofID      uuid.UUID              `json:"proof_id"`
	UserID       uuid.UUID              `json:"user_id"`
	ProofSystem  models.ProofSystemType `json:"proof_system"`
	Data         *models.InputData      `json:"data"`
	PublicInputs json.RawMessage        `json:"public_inputs,omitempty"`
	CircuitID    *uuid.UUID             `json:"circuit_id,omitempty"`
	TemplateID   *uuid.UUID             `json:"template_id,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// EnqueueProofGeneration enqueues a proof generation job
func (c *Client) EnqueueProofGeneration(ctx context.Context, payload *ProofGenerationPayload, priority int) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeProofGeneration, data)

	opts := []asynq.Option{
		asynq.Queue("proofs"),
		asynq.MaxRetry(3),
		asynq.Timeout(10 * time.Minute),
	}

	// Set priority based on user tier or explicit priority
	switch priority {
	case 1:
		opts = append(opts, asynq.Queue("proofs:high"))
	case -1:
		opts = append(opts, asynq.Queue("proofs:low"))
	}

	info, err := c.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Printf("Enqueued proof generation job: %s (queue: %s)\n", info.ID, info.Queue)
	return nil
}

// Server wraps an asynq server for processing jobs
type Server struct {
	server *asynq.Server
}

// NewServer creates a new queue server
func NewServer(redisAddr, password string, concurrency int) *Server {
	opt := asynq.RedisClientOpt{Addr: redisAddr}
	if password != "" {
		opt.Password = password
	}

	server := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"proofs:high": 6, // High priority
				"proofs":      3, // Normal priority
				"proofs:low":  1, // Low priority
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				fmt.Printf("Task %s failed: %v\n", task.Type(), err)
			}),
		},
	)

	return &Server{server: server}
}

// Start starts the queue server
func (s *Server) Start(mux *asynq.ServeMux) error {
	return s.server.Start(mux)
}

// Stop stops the queue server
func (s *Server) Stop() {
	s.server.Stop()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.server.Shutdown()
}

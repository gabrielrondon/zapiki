package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProofSystemType represents the type of proof system
type ProofSystemType string

const (
	ProofSystemCommitment ProofSystemType = "commitment"
	ProofSystemGroth16    ProofSystemType = "groth16"
	ProofSystemPLONK      ProofSystemType = "plonk"
	ProofSystemSTARK      ProofSystemType = "stark"
)

// ProofStatus represents the status of a proof generation job
type ProofStatus string

const (
	ProofStatusPending   ProofStatus = "pending"
	ProofStatusProcessing ProofStatus = "processing"
	ProofStatusCompleted ProofStatus = "completed"
	ProofStatusFailed    ProofStatus = "failed"
)

// DataType represents the type of input data
type DataType string

const (
	DataTypeString DataType = "string"
	DataTypeJSON   DataType = "json"
	DataTypeBytes  DataType = "bytes"
)

// User represents a user account
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Tier      string    `json:"tier" db:"tier"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Key         string    `json:"key" db:"key"`
	Name        string    `json:"name" db:"name"`
	RateLimit   int       `json:"rate_limit" db:"rate_limit"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
}

// Circuit represents a circuit definition
type Circuit struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	UserID             uuid.UUID       `json:"user_id" db:"user_id"`
	Name               string          `json:"name" db:"name"`
	Description        string          `json:"description" db:"description"`
	ProofSystem        ProofSystemType `json:"proof_system" db:"proof_system"`
	CircuitDefinition  json.RawMessage `json:"circuit_definition" db:"circuit_definition"`
	ProvingKeyURL      string          `json:"proving_key_url,omitempty" db:"proving_key_url"`
	VerificationKeyURL string          `json:"verification_key_url,omitempty" db:"verification_key_url"`
	IsPublic           bool            `json:"is_public" db:"is_public"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// Proof represents a proof record
type Proof struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	UserID        uuid.UUID       `json:"user_id" db:"user_id"`
	CircuitID     *uuid.UUID      `json:"circuit_id,omitempty" db:"circuit_id"`
	TemplateID    *uuid.UUID      `json:"template_id,omitempty" db:"template_id"`
	ProofSystem   ProofSystemType `json:"proof_system" db:"proof_system"`
	Status        ProofStatus     `json:"status" db:"status"`
	InputData     json.RawMessage `json:"input_data,omitempty" db:"input_data"`
	ProofData     json.RawMessage `json:"proof_data,omitempty" db:"proof_data"`
	PublicInputs  json.RawMessage `json:"public_inputs,omitempty" db:"public_inputs"`
	ProofURL      string          `json:"proof_url,omitempty" db:"proof_url"`
	ErrorMessage  string          `json:"error_message,omitempty" db:"error_message"`
	GenerationTimeMs int64        `json:"generation_time_ms,omitempty" db:"generation_time_ms"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
}

// Verification represents a verification record
type Verification struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	ProofID      uuid.UUID       `json:"proof_id" db:"proof_id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	ProofSystem  ProofSystemType `json:"proof_system" db:"proof_system"`
	IsValid      bool            `json:"is_valid" db:"is_valid"`
	ErrorMessage string          `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// Template represents a pre-built circuit template
type Template struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Description     string          `json:"description" db:"description"`
	Category        string          `json:"category" db:"category"`
	ProofSystem     ProofSystemType `json:"proof_system" db:"proof_system"`
	CircuitID       uuid.UUID       `json:"circuit_id" db:"circuit_id"`
	InputSchema     json.RawMessage `json:"input_schema" db:"input_schema"`
	ExampleInputs   json.RawMessage `json:"example_inputs" db:"example_inputs"`
	Documentation   string          `json:"documentation" db:"documentation"`
	IsActive        bool            `json:"is_active" db:"is_active"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// Job represents an async job
type Job struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	ProofID     uuid.UUID       `json:"proof_id" db:"proof_id"`
	Status      ProofStatus     `json:"status" db:"status"`
	Priority    int             `json:"priority" db:"priority"`
	RetryCount  int             `json:"retry_count" db:"retry_count"`
	MaxRetries  int             `json:"max_retries" db:"max_retries"`
	ErrorMessage string         `json:"error_message,omitempty" db:"error_message"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
}

// UsageMetric represents usage analytics
type UsageMetric struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	ProofSystem ProofSystemType `json:"proof_system" db:"proof_system"`
	Operation   string          `json:"operation" db:"operation"`
	Success     bool            `json:"success" db:"success"`
	DurationMs  int64           `json:"duration_ms" db:"duration_ms"`
	Timestamp   time.Time       `json:"timestamp" db:"timestamp"`
}

// InputData represents the input data for proof generation
type InputData struct {
	Type  DataType        `json:"type"`
	Value json.RawMessage `json:"value"`
}

// ProofOptions represents options for proof generation
type ProofOptions struct {
	TemplateID *uuid.UUID `json:"template_id,omitempty"`
	CircuitID  *uuid.UUID `json:"circuit_id,omitempty"`
	Async      bool       `json:"async,omitempty"`
}

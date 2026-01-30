package prover

import (
	"context"
	"encoding/json"

	"github.com/gabrielrondon/zapiki/internal/models"
)

// ProofSystem defines the interface that all proof systems must implement
type ProofSystem interface {
	// Name returns the name of the proof system
	Name() models.ProofSystemType

	// Setup performs any required setup for the circuit (e.g., trusted setup)
	// For commitment-based proofs, this is typically a no-op
	Setup(ctx context.Context, circuit *models.Circuit) (*SetupResult, error)

	// Generate creates a proof from the given request
	Generate(ctx context.Context, req *ProofRequest) (*ProofResponse, error)

	// Verify checks if a proof is valid
	Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error)

	// Capabilities returns the capabilities of this proof system
	Capabilities() Capabilities
}

// SetupResult contains the results of a circuit setup
type SetupResult struct {
	ProvingKey      json.RawMessage `json:"proving_key"`
	VerificationKey json.RawMessage `json:"verification_key"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ProofRequest represents a request to generate a proof
type ProofRequest struct {
	Circuit      *models.Circuit    `json:"circuit,omitempty"`
	Data         *models.InputData  `json:"data"`
	PublicInputs json.RawMessage    `json:"public_inputs,omitempty"`
	ProvingKey   json.RawMessage    `json:"proving_key,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// ProofResponse contains the generated proof
type ProofResponse struct {
	Proof            json.RawMessage `json:"proof"`
	PublicInputs     json.RawMessage `json:"public_inputs,omitempty"`
	VerificationKey  json.RawMessage `json:"verification_key,omitempty"`
	GenerationTimeMs int64           `json:"generation_time_ms"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// VerifyRequest represents a request to verify a proof
type VerifyRequest struct {
	Proof           json.RawMessage `json:"proof"`
	VerificationKey json.RawMessage `json:"verification_key"`
	PublicInputs    json.RawMessage `json:"public_inputs,omitempty"`
}

// VerifyResponse contains the verification result
type VerifyResponse struct {
	Valid        bool   `json:"valid"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Capabilities describes what a proof system can do
type Capabilities struct {
	// SupportsSetup indicates if the system requires a setup phase
	SupportsSetup bool `json:"supports_setup"`

	// RequiresTrustedSetup indicates if setup requires trusted parties
	RequiresTrustedSetup bool `json:"requires_trusted_setup"`

	// SupportsCustomCircuits indicates if users can define custom circuits
	SupportsCustomCircuits bool `json:"supports_custom_circuits"`

	// AsyncOnly indicates if proof generation must be async
	AsyncOnly bool `json:"async_only"`

	// TypicalGenerationTime is the typical time to generate a proof (ms)
	TypicalGenerationTime int64 `json:"typical_generation_time"`

	// MaxProofSize is the maximum proof size in bytes
	MaxProofSize int64 `json:"max_proof_size"`

	// Features lists specific features (e.g., "zero-knowledge", "post-quantum")
	Features []string `json:"features"`
}

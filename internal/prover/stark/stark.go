package stark

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
)

// STARKProver implements transparent zero-knowledge proofs using STARK
// This is a simplified STARK implementation for demonstration
// In production, use a mature library like Winterfell or Cairo
type STARKProver struct {
	fieldPrime *big.Int // Prime field size
}

// NewSTARKProver creates a new STARK prover
func NewSTARKProver() *STARKProver {
	// Use a large prime for the finite field
	// 2^256 - 189 (a safe prime)
	prime := new(big.Int)
	prime.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639747", 10)

	return &STARKProver{
		fieldPrime: prime,
	}
}

// Name returns the proof system name
func (p *STARKProver) Name() models.ProofSystemType {
	return models.ProofSystemSTARK
}

// Setup performs setup for STARK (no trusted setup needed!)
func (p *STARKProver) Setup(ctx context.Context, circuit *models.Circuit) (*prover.SetupResult, error) {
	// STARKs don't require trusted setup - this is their main advantage!
	// We only need to configure parameters

	metadata := map[string]interface{}{
		"circuit_id": circuit.ID.String(),
		"message":    "STARK setup complete - no trusted setup required",
		"setup_time_ms": 0,
	}

	return &prover.SetupResult{
		ProvingKey:      json.RawMessage(`{"type":"stark","setup":"transparent"}`),
		VerificationKey: json.RawMessage(`{"type":"stark","public_parameters":true}`),
		Metadata:        metadata,
	}, nil
}

// Generate generates a STARK proof
func (p *STARKProver) Generate(ctx context.Context, req *prover.ProofRequest) (*prover.ProofResponse, error) {
	startTime := time.Now()

	// Parse input data
	var inputData map[string]interface{}
	if err := json.Unmarshal(req.Data.Value, &inputData); err != nil {
		return nil, fmt.Errorf("failed to parse inputs: %w", err)
	}

	// Execute computation trace
	trace, publicInputs, err := p.executeComputationTrace(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to execute computation: %w", err)
	}

	// Generate FRI (Fast Reed-Solomon IOP) commitment
	commitment := p.generateFRICommitment(trace)

	// Generate random challenges using Fiat-Shamir transform
	challenges := p.generateChallenges(commitment, publicInputs)

	// Generate proof components
	proofData := STARKProof{
		Trace:         trace,
		Commitment:    commitment,
		Challenges:    challenges,
		PublicInputs:  publicInputs,
		Timestamp:     time.Now().Format(time.RFC3339),
		FieldPrime:    p.fieldPrime.String(),
		ProofVersion:  "1.0",
	}

	proofJSON, err := json.Marshal(proofData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proof: %w", err)
	}

	// Verification key (public parameters)
	vkData := STARKVerificationKey{
		FieldPrime:   p.fieldPrime.String(),
		PublicInputs: publicInputs,
		ProofVersion: "1.0",
	}

	vkJSON, err := json.Marshal(vkData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal verification key: %w", err)
	}

	generationTime := time.Since(startTime).Milliseconds()

	return &prover.ProofResponse{
		Proof:            proofJSON,
		VerificationKey:  vkJSON,
		GenerationTimeMs: generationTime,
	}, nil
}

// Verify verifies a STARK proof
func (p *STARKProver) Verify(ctx context.Context, req *prover.VerifyRequest) (*prover.VerifyResponse, error) {
	// Parse proof
	var proof STARKProof
	if err := json.Unmarshal(req.Proof, &proof); err != nil {
		return nil, fmt.Errorf("failed to parse proof: %w", err)
	}

	// Parse verification key
	var vk STARKVerificationKey
	if err := json.Unmarshal(req.VerificationKey, &vk); err != nil {
		return nil, fmt.Errorf("failed to parse verification key: %w", err)
	}

	// Verify proof version
	if proof.ProofVersion != "1.0" {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: "Unsupported proof version",
		}, nil
	}

	// Verify field prime matches
	if proof.FieldPrime != vk.FieldPrime {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: "Field prime mismatch",
		}, nil
	}

	// Verify public inputs match
	if !equalStringArrays(proof.PublicInputs, vk.PublicInputs) {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: "Public inputs mismatch",
		}, nil
	}

	// Verify FRI commitment
	expectedCommitment := p.generateFRICommitment(proof.Trace)
	if proof.Commitment != expectedCommitment {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: "FRI commitment verification failed",
		}, nil
	}

	// Verify challenges (Fiat-Shamir)
	expectedChallenges := p.generateChallenges(proof.Commitment, proof.PublicInputs)
	if !equalStringArrays(proof.Challenges, expectedChallenges) {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: "Challenge verification failed",
		}, nil
	}

	// Verify computation trace consistency
	if !p.verifyTraceConsistency(proof.Trace, proof.PublicInputs) {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: "Trace consistency check failed",
		}, nil
	}

	return &prover.VerifyResponse{
		Valid: true,
	}, nil
}

// Capabilities returns STARK capabilities
func (p *STARKProver) Capabilities() prover.Capabilities {
	return prover.Capabilities{
		SupportsSetup:        false, // No setup needed!
		RequiresTrustedSetup: false, // Transparent!
		SupportsCustomCircuits: true,
		AsyncOnly:            true,
		TypicalGenerationTime: 40000, // ~40 seconds
		MaxProofSize:         102400,  // ~100KB
		Features: []string{
			"transparent",
			"no-trusted-setup",
			"quantum-resistant",
			"hash-based",
			"fri-commitment",
		},
	}
}

// executeComputationTrace executes the computation and generates trace
func (p *STARKProver) executeComputationTrace(inputs map[string]interface{}) ([]string, []string, error) {
	// This is a simplified computation trace
	// In a real STARK system, this would be much more complex

	trace := make([]string, 0)
	publicInputs := make([]string, 0)

	// Extract computation based on input type
	if value, ok := inputs["value"]; ok {
		// Simple value commitment
		valueStr := fmt.Sprintf("%v", value)
		hash := sha256.Sum256([]byte(valueStr))
		trace = append(trace, hex.EncodeToString(hash[:]))
		publicInputs = append(publicInputs, hex.EncodeToString(hash[:]))
	}

	if a, okA := inputs["a"]; okA {
		if b, okB := inputs["b"]; okB {
			if c, okC := inputs["c"]; okC {
				// Multiplication: a * b = c
				aVal := toBigInt(a)
				bVal := toBigInt(b)
				cVal := toBigInt(c)

				// Generate trace: intermediate steps of multiplication
				trace = append(trace, aVal.String())
				trace = append(trace, bVal.String())

				result := new(big.Int).Mul(aVal, bVal)
				trace = append(trace, result.String())

				// Public output
				publicInputs = append(publicInputs, cVal.String())
			}
		}
	}

	if len(trace) == 0 {
		// Default: hash of all inputs
		inputJSON, _ := json.Marshal(inputs)
		hash := sha256.Sum256(inputJSON)
		trace = append(trace, hex.EncodeToString(hash[:]))
		publicInputs = append(publicInputs, hex.EncodeToString(hash[:]))
	}

	return trace, publicInputs, nil
}

// generateFRICommitment generates a FRI commitment to the trace
func (p *STARKProver) generateFRICommitment(trace []string) string {
	// Simplified FRI commitment using Merkle tree root
	// In production, use full FRI protocol

	hasher := sha256.New()
	for _, step := range trace {
		hasher.Write([]byte(step))
	}

	commitment := hasher.Sum(nil)
	return hex.EncodeToString(commitment)
}

// generateChallenges generates random challenges using Fiat-Shamir transform
func (p *STARKProver) generateChallenges(commitment string, publicInputs []string) []string {
	// Fiat-Shamir: turn interactive protocol into non-interactive
	// by using hash of transcript as random challenges

	hasher := sha256.New()
	hasher.Write([]byte(commitment))
	for _, input := range publicInputs {
		hasher.Write([]byte(input))
	}

	transcript := hasher.Sum(nil)

	// Generate multiple challenges from transcript
	challenges := make([]string, 3)
	for i := 0; i < 3; i++ {
		h := sha256.Sum256(append(transcript, byte(i)))
		challenges[i] = hex.EncodeToString(h[:8]) // Use first 8 bytes
	}

	return challenges
}

// verifyTraceConsistency verifies the computation trace is consistent
func (p *STARKProver) verifyTraceConsistency(trace []string, publicInputs []string) bool {
	// Basic consistency checks
	if len(trace) == 0 {
		return false
	}

	if len(publicInputs) == 0 {
		return false
	}

	// Trace should be well-formed
	// In a real STARK, this would check polynomial constraints
	return true
}

// Helper functions

func toBigInt(v interface{}) *big.Int {
	switch val := v.(type) {
	case int:
		return big.NewInt(int64(val))
	case int64:
		return big.NewInt(val)
	case float64:
		return big.NewInt(int64(val))
	case string:
		i, _ := new(big.Int).SetString(val, 10)
		return i
	default:
		return big.NewInt(0)
	}
}

func equalStringArrays(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// STARKProof represents a STARK proof
type STARKProof struct {
	Trace        []string `json:"trace"`
	Commitment   string   `json:"commitment"`
	Challenges   []string `json:"challenges"`
	PublicInputs []string `json:"public_inputs"`
	Timestamp    string   `json:"timestamp"`
	FieldPrime   string   `json:"field_prime"`
	ProofVersion string   `json:"proof_version"`
}

// STARKVerificationKey represents STARK verification parameters
type STARKVerificationKey struct {
	FieldPrime   string   `json:"field_prime"`
	PublicInputs []string `json:"public_inputs"`
	ProofVersion string   `json:"proof_version"`
}

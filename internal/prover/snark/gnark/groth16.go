package gnark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
)

// Groth16Prover implements Groth16 SNARK proof system
type Groth16Prover struct {
	curve ecc.ID
}

// NewGroth16Prover creates a new Groth16 prover
func NewGroth16Prover() *Groth16Prover {
	return &Groth16Prover{
		curve: ecc.BN254, // BN254 curve (widely used, ~128-bit security)
	}
}

// Name returns the proof system name
func (p *Groth16Prover) Name() models.ProofSystemType {
	return models.ProofSystemGroth16
}

// Setup performs trusted setup for a circuit
func (p *Groth16Prover) Setup(ctx context.Context, circuit *models.Circuit) (*prover.SetupResult, error) {
	// Parse circuit definition to get circuit type
	var circuitDef struct {
		CircuitType string                 `json:"circuit_type"`
		Params      map[string]interface{} `json:"params"`
	}

	if err := json.Unmarshal(circuit.CircuitDefinition, &circuitDef); err != nil {
		return nil, fmt.Errorf("failed to parse circuit definition: %w", err)
	}

	// Get circuit instance
	circuitInstance, err := GetCircuitByName(circuitDef.CircuitType)
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit: %w", err)
	}

	// Compile circuit to R1CS
	ccs, err := frontend.Compile(p.curve.ScalarField(), r1cs.NewBuilder, circuitInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %w", err)
	}

	// Run trusted setup
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		return nil, fmt.Errorf("failed to setup: %w", err)
	}

	// Serialize keys
	pkBuf := new(bytes.Buffer)
	if _, err := pk.WriteTo(pkBuf); err != nil {
		return nil, fmt.Errorf("failed to serialize proving key: %w", err)
	}

	vkBuf := new(bytes.Buffer)
	if _, err := vk.WriteTo(vkBuf); err != nil {
		return nil, fmt.Errorf("failed to serialize verification key: %w", err)
	}

	return &prover.SetupResult{
		ProvingKey:      pkBuf.Bytes(),
		VerificationKey: vkBuf.Bytes(),
		Metadata: map[string]interface{}{
			"curve":       p.curve.String(),
			"constraints": ccs.GetNbConstraints(),
			"variables":   ccs.GetNbSecretVariables() + ccs.GetNbPublicVariables(),
		},
	}, nil
}

// Generate creates a Groth16 proof
func (p *Groth16Prover) Generate(ctx context.Context, req *prover.ProofRequest) (*prover.ProofResponse, error) {
	startTime := time.Now()

	// Parse circuit and input data
	var circuitDef struct {
		CircuitType string                 `json:"circuit_type"`
		Params      map[string]interface{} `json:"params"`
	}

	if req.Circuit != nil && req.Circuit.CircuitDefinition != nil {
		if err := json.Unmarshal(req.Circuit.CircuitDefinition, &circuitDef); err != nil {
			return nil, fmt.Errorf("failed to parse circuit definition: %w", err)
		}
	} else {
		// Default to simple circuit for demo
		circuitDef.CircuitType = "simple"
	}

	// Parse input data
	var inputData map[string]interface{}
	if err := json.Unmarshal(req.Data.Value, &inputData); err != nil {
		return nil, fmt.Errorf("failed to parse input data: %w", err)
	}

	// Get circuit instance and assign values
	witness, err := p.createWitness(circuitDef.CircuitType, inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %w", err)
	}

	// Compile circuit
	circuitInstance, _ := GetCircuitByName(circuitDef.CircuitType)
	ccs, err := frontend.Compile(p.curve.ScalarField(), r1cs.NewBuilder, circuitInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %w", err)
	}

	// Check if we have proving key or need to generate it
	var pk groth16.ProvingKey
	var vk groth16.VerifyingKey

	if req.ProvingKey != nil && len(req.ProvingKey) > 0 {
		// Deserialize existing proving key
		pkBuf := bytes.NewReader(req.ProvingKey)
		pk = groth16.NewProvingKey(p.curve)
		if _, err := pk.ReadFrom(pkBuf); err != nil {
			return nil, fmt.Errorf("failed to deserialize proving key: %w", err)
		}

		// We need vk too, run setup again (in real implementation, cache this)
		_, vk, _ = groth16.Setup(ccs)
	} else {
		// Run setup (in production, this should be done once and cached)
		pk, vk, err = groth16.Setup(ccs)
		if err != nil {
			return nil, fmt.Errorf("failed to setup: %w", err)
		}
	}

	// Generate witness
	fullWitness, err := frontend.NewWitness(witness, p.curve.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %w", err)
	}

	// Generate proof
	proof, err := groth16.Prove(ccs, pk, fullWitness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	// Serialize proof
	proofBuf := new(bytes.Buffer)
	if _, err := proof.WriteTo(proofBuf); err != nil {
		return nil, fmt.Errorf("failed to serialize proof: %w", err)
	}

	// Serialize verification key
	vkBuf := new(bytes.Buffer)
	if _, err := vk.WriteTo(vkBuf); err != nil {
		return nil, fmt.Errorf("failed to serialize verification key: %w", err)
	}

	// Extract public inputs
	publicWitness, err := fullWitness.Public()
	if err != nil {
		return nil, fmt.Errorf("failed to extract public inputs: %w", err)
	}

	publicBuf := new(bytes.Buffer)
	if _, err := publicWitness.WriteTo(publicBuf); err != nil {
		return nil, fmt.Errorf("failed to serialize public inputs: %w", err)
	}

	generationTime := time.Since(startTime).Milliseconds()

	return &prover.ProofResponse{
		Proof:            proofBuf.Bytes(),
		PublicInputs:     publicBuf.Bytes(),
		VerificationKey:  vkBuf.Bytes(),
		GenerationTimeMs: generationTime,
		Metadata: map[string]interface{}{
			"proof_system": "groth16",
			"curve":        p.curve.String(),
			"circuit_type": circuitDef.CircuitType,
		},
	}, nil
}

// Verify verifies a Groth16 proof
func (p *Groth16Prover) Verify(ctx context.Context, req *prover.VerifyRequest) (*prover.VerifyResponse, error) {
	// Deserialize verification key
	vkBuf := bytes.NewReader(req.VerificationKey)
	vk := groth16.NewVerifyingKey(p.curve)
	if _, err := vk.ReadFrom(vkBuf); err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to deserialize verification key: %v", err),
		}, nil
	}

	// Deserialize proof
	proofBuf := bytes.NewReader(req.Proof)
	proof := groth16.NewProof(p.curve)
	if _, err := proof.ReadFrom(proofBuf); err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to deserialize proof: %v", err),
		}, nil
	}

	// Deserialize public inputs
	publicBuf := bytes.NewReader(req.PublicInputs)
	publicWitness, err := frontend.NewWitness(nil, p.curve.ScalarField())
	if err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to create witness: %v", err),
		}, nil
	}

	if _, err := publicWitness.ReadFrom(publicBuf); err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to deserialize public inputs: %v", err),
		}, nil
	}

	// Verify proof
	err = groth16.Verify(proof, vk, publicWitness)
	if err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("verification failed: %v", err),
		}, nil
	}

	return &prover.VerifyResponse{
		Valid: true,
	}, nil
}

// Capabilities returns Groth16 capabilities
func (p *Groth16Prover) Capabilities() prover.Capabilities {
	return prover.Capabilities{
		SupportsSetup:          true,
		RequiresTrustedSetup:   true,
		SupportsCustomCircuits: true,
		AsyncOnly:              true, // Groth16 proof generation can take seconds
		TypicalGenerationTime:  30000, // ~30 seconds for medium circuits
		MaxProofSize:           1024,  // ~1KB proof size
		Features: []string{
			"zero-knowledge",
			"succinct-proofs",
			"fast-verification",
			"trusted-setup-required",
		},
	}
}

// createWitness creates a witness from input data
func (p *Groth16Prover) createWitness(circuitType string, inputData map[string]interface{}) (frontend.Circuit, error) {
	switch circuitType {
	case "simple":
		return &SimpleCircuit{
			X: inputData["x"],
			Y: inputData["y"],
			Z: inputData["z"],
		}, nil

	case "age_verification":
		return &AgeVerificationCircuit{
			Age:     inputData["age"],
			MinAge:  inputData["min_age"],
			IsAdult: inputData["is_adult"],
		}, nil

	case "range_proof":
		return &RangeProofCircuit{
			Value:   inputData["value"],
			Min:     inputData["min"],
			Max:     inputData["max"],
			InRange: inputData["in_range"],
		}, nil

	default:
		return &SimpleCircuit{
			X: 3,
			Y: 5,
			Z: 15,
		}, nil
	}
}

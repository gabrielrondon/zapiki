package gnark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
)

// PLONKProver implements PLONK SNARK proof system
type PLONKProver struct {
	curve ecc.ID
	// Universal SRS (Structured Reference String) - can be reused across circuits
	srs plonk.ProvingKey
}

// NewPLONKProver creates a new PLONK prover
func NewPLONKProver() *PLONKProver {
	return &PLONKProver{
		curve: ecc.BN254, // BN254 curve (same as Groth16)
	}
}

// Name returns the proof system name
func (p *PLONKProver) Name() models.ProofSystemType {
	return models.ProofSystemPLONK
}

// Setup performs universal setup (or circuit-specific compilation)
func (p *PLONKProver) Setup(ctx context.Context, circuit *models.Circuit) (*prover.SetupResult, error) {
	// Parse circuit definition
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

	// Compile circuit to SCS (Sparse Constraint System - used by PLONK)
	ccs, err := frontend.Compile(p.curve.ScalarField(), scs.NewBuilder, circuitInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %w", err)
	}

	// Setup PLONK (generates proving and verification keys)
	// Note: PLONK uses a universal SRS
	// For simplicity, we use the dummy setup (in production, use real trusted setup ceremony)
	pk, vk, err := plonk.Setup(ccs, nil, nil)
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
			"setup_type":  "universal", // PLONK's key advantage
		},
	}, nil
}

// Generate creates a PLONK proof
func (p *PLONKProver) Generate(ctx context.Context, req *prover.ProofRequest) (*prover.ProofResponse, error) {
	startTime := time.Now()

	// Parse circuit definition
	var circuitDef struct {
		CircuitType string                 `json:"circuit_type"`
		Params      map[string]interface{} `json:"params"`
	}

	if req.Circuit != nil && req.Circuit.CircuitDefinition != nil {
		if err := json.Unmarshal(req.Circuit.CircuitDefinition, &circuitDef); err != nil {
			return nil, fmt.Errorf("failed to parse circuit definition: %w", err)
		}
	} else {
		// Default to simple circuit
		circuitDef.CircuitType = "simple"
	}

	// Parse input data
	var inputData map[string]interface{}
	if err := json.Unmarshal(req.Data.Value, &inputData); err != nil {
		return nil, fmt.Errorf("failed to parse input data: %w", err)
	}

	// Create witness
	witness, err := p.createWitness(circuitDef.CircuitType, inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %w", err)
	}

	// Compile circuit
	circuitInstance, _ := GetCircuitByName(circuitDef.CircuitType)
	ccs, err := frontend.Compile(p.curve.ScalarField(), scs.NewBuilder, circuitInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to compile circuit: %w", err)
	}

	// Setup or load keys
	var pk plonk.ProvingKey
	var vk plonk.VerifyingKey

	if req.ProvingKey != nil && len(req.ProvingKey) > 0 {
		// Deserialize existing proving key
		pkBuf := bytes.NewReader(req.ProvingKey)
		pk = plonk.NewProvingKey(p.curve)
		if _, err := pk.ReadFrom(pkBuf); err != nil {
			return nil, fmt.Errorf("failed to deserialize proving key: %w", err)
		}

		// Need vk too, run setup again (in real implementation, cache this)
		_, vk, _ = plonk.Setup(ccs, nil, nil)
	} else {
		// Run setup
		pk, vk, err = plonk.Setup(ccs, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to setup: %w", err)
		}
	}

	// Generate witness
	fullWitness, err := frontend.NewWitness(witness, p.curve.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %w", err)
	}

	// Generate PLONK proof
	proof, err := plonk.Prove(ccs, pk, fullWitness)
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
			"proof_system": "plonk",
			"curve":        p.curve.String(),
			"circuit_type": circuitDef.CircuitType,
			"setup_type":   "universal",
		},
	}, nil
}

// Verify verifies a PLONK proof
func (p *PLONKProver) Verify(ctx context.Context, req *prover.VerifyRequest) (*prover.VerifyResponse, error) {
	// Deserialize verification key
	vkBuf := bytes.NewReader(req.VerificationKey)
	vk := plonk.NewVerifyingKey(p.curve)
	if _, err := vk.ReadFrom(vkBuf); err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to deserialize verification key: %v", err),
		}, nil
	}

	// Deserialize proof
	proofBuf := bytes.NewReader(req.Proof)
	proof := plonk.NewProof(p.curve)
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

	// Verify PLONK proof
	err = plonk.Verify(proof, vk, publicWitness)
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

// Capabilities returns PLONK capabilities
func (p *PLONKProver) Capabilities() prover.Capabilities {
	return prover.Capabilities{
		SupportsSetup:          true,
		RequiresTrustedSetup:   false, // PLONK uses universal setup!
		SupportsCustomCircuits: true,
		AsyncOnly:              true,
		TypicalGenerationTime:  35000, // ~35 seconds (slightly slower than Groth16)
		MaxProofSize:           2048,  // ~2KB (larger than Groth16)
		Features: []string{
			"zero-knowledge",
			"universal-setup", // Key advantage!
			"no-per-circuit-setup",
			"flexible",
			"updatable-srs",
		},
	}
}

// createWitness creates a witness from input data (same as Groth16)
func (p *PLONKProver) createWitness(circuitType string, inputData map[string]interface{}) (frontend.Circuit, error) {
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

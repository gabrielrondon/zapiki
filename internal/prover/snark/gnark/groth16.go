package gnark

import (
	"bytes"
	"context"
	"encoding/base64"
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

	// Check for circuit_type in Options first (new method)
	if req.Options != nil {
		if ct, ok := req.Options["circuit_type"].(string); ok {
			circuitDef.CircuitType = ct
		}
	}

	// Fallback to Circuit.CircuitDefinition (old method)
	if circuitDef.CircuitType == "" && req.Circuit != nil && req.Circuit.CircuitDefinition != nil {
		if err := json.Unmarshal(req.Circuit.CircuitDefinition, &circuitDef); err != nil {
			return nil, fmt.Errorf("failed to parse circuit definition: %w", err)
		}
	}

	// Parse input data first (needed for auto-detection)
	var inputData map[string]interface{}
	if err := json.Unmarshal(req.Data.Value, &inputData); err != nil {
		return nil, fmt.Errorf("failed to parse input data: %w", err)
	}

	// Auto-detect circuit type based on input fields if not specified
	if circuitDef.CircuitType == "" {
		circuitDef.CircuitType = detectCircuitType(inputData)
	}

	// Default to simple circuit if still not detected
	if circuitDef.CircuitType == "" {
		circuitDef.CircuitType = "simple"
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

	// Encode binary data as base64-encoded JSON for database storage
	proofJSON, _ := json.Marshal(map[string]string{
		"proof": base64.StdEncoding.EncodeToString(proofBuf.Bytes()),
	})
	publicInputsJSON, _ := json.Marshal(map[string]string{
		"public_inputs": base64.StdEncoding.EncodeToString(publicBuf.Bytes()),
	})
	vkJSON, _ := json.Marshal(map[string]string{
		"verification_key": base64.StdEncoding.EncodeToString(vkBuf.Bytes()),
	})

	return &prover.ProofResponse{
		Proof:            json.RawMessage(proofJSON),
		PublicInputs:     json.RawMessage(publicInputsJSON),
		VerificationKey:  json.RawMessage(vkJSON),
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
// Helper function to convert interface{} to int, handling both float64 and int
func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	default:
		return 0
	}
}

func (p *Groth16Prover) createWitness(circuitType string, inputData map[string]interface{}) (frontend.Circuit, error) {
	switch circuitType {
	case "simple":
		return &SimpleCircuit{
			X: toInt(inputData["x"]),
			Y: toInt(inputData["y"]),
			Z: toInt(inputData["z"]),
		}, nil

	case "age_verification":
		return &AgeVerificationCircuit{
			Age:     toInt(inputData["age"]),
			MinAge:  toInt(inputData["min_age"]),
			IsAdult: toInt(inputData["is_adult"]),
		}, nil

	case "range_proof":
		return &RangeProofCircuit{
			Value:   toInt(inputData["value"]),
			Min:     toInt(inputData["min"]),
			Max:     toInt(inputData["max"]),
			InRange: toInt(inputData["in_range"]),
		}, nil

	// AML/KYC Compliance circuits
	case "aml_age_verification":
		return &AMLAgeVerificationCircuit{
			MinimumAge:  toInt(inputData["minimum_age"]),
			CurrentYear: toInt(inputData["current_year"]),
			BirthYear:   toInt(inputData["birth_year"]),
			Nonce:       toInt(inputData["nonce"]),
		}, nil

	case "aml_sanctions_check":
		return &AMLSanctionsCheckCircuit{
			SanctionsListRoot: toInt(inputData["sanctions_list_root"]),
			CurrentTimestamp:  toInt(inputData["current_timestamp"]),
			UserIdentifier:    toInt(inputData["user_identifier"]),
		}, nil

	case "aml_residency_proof":
		return &AMLResidencyProofCircuit{
			AllowedCountryCode: toInt(inputData["allowed_country_code"]),
			CurrentTimestamp:   toInt(inputData["current_timestamp"]),
			UserCountryCode:    toInt(inputData["user_country_code"]),
			AddressHash:        toInt(inputData["address_hash"]),
		}, nil

	case "aml_income_verification":
		return &AMLIncomeVerificationCircuit{
			MinimumIncome:    toInt(inputData["minimum_income"]),
			CurrentTimestamp: toInt(inputData["current_timestamp"]),
			ActualIncome:     toInt(inputData["actual_income"]),
			IncomeSourceHash: toInt(inputData["income_source_hash"]),
		}, nil

	default:
		return &SimpleCircuit{
			X: 3,
			Y: 5,
			Z: 15,
		}, nil
	}
}

// detectCircuitType automatically detects the circuit type based on input fields
func detectCircuitType(inputData map[string]interface{}) string {
	// AML Age Verification: has minimum_age, current_year, birth_year
	if _, hasMinAge := inputData["minimum_age"]; hasMinAge {
		if _, hasCurYear := inputData["current_year"]; hasCurYear {
			if _, hasBirthYear := inputData["birth_year"]; hasBirthYear {
				return "aml_age_verification"
			}
		}
	}

	// AML Sanctions Check: has sanctions_list_root, user_identifier
	if _, hasSanctions := inputData["sanctions_list_root"]; hasSanctions {
		if _, hasUser := inputData["user_identifier"]; hasUser {
			return "aml_sanctions_check"
		}
	}

	// AML Residency: has allowed_country_code, user_country_code
	if _, hasAllowed := inputData["allowed_country_code"]; hasAllowed {
		if _, hasUser := inputData["user_country_code"]; hasUser {
			return "aml_residency_proof"
		}
	}

	// AML Income: has minimum_income, actual_income
	if _, hasMinIncome := inputData["minimum_income"]; hasMinIncome {
		if _, hasActual := inputData["actual_income"]; hasActual {
			return "aml_income_verification"
		}
	}

	// Old age verification: has age, min_age, is_adult
	if _, hasAge := inputData["age"]; hasAge {
		if _, hasMinAge := inputData["min_age"]; hasMinAge {
			return "age_verification"
		}
	}

	// Range proof: has value, min, max
	if _, hasValue := inputData["value"]; hasValue {
		if _, hasMin := inputData["min"]; hasMin {
			if _, hasMax := inputData["max"]; hasMax {
				return "range_proof"
			}
		}
	}

	// Simple circuit: has x, y, z
	if _, hasX := inputData["x"]; hasX {
		if _, hasY := inputData["y"]; hasY {
			if _, hasZ := inputData["z"]; hasZ {
				return "simple"
			}
		}
	}

	// Unknown - return empty string
	return ""
}

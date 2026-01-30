package gnark

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
)

func TestGroth16Prover_SimpleCircuit(t *testing.T) {
	p := NewGroth16Prover()

	// Test data: x=3, y=5, z=15 (3*5=15)
	inputData := map[string]interface{}{
		"x": 3,
		"y": 5,
		"z": 15,
	}
	inputJSON, _ := json.Marshal(inputData)

	circuitDef := map[string]interface{}{
		"circuit_type": "simple",
	}
	circuitDefJSON, _ := json.Marshal(circuitDef)

	circuit := &models.Circuit{
		CircuitDefinition: circuitDefJSON,
	}

	req := &prover.ProofRequest{
		Circuit: circuit,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: inputJSON,
		},
	}

	ctx := context.Background()

	// Generate proof
	t.Log("Generating Groth16 proof...")
	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	if resp.Proof == nil {
		t.Error("Expected proof to be non-nil")
	}

	if resp.GenerationTimeMs <= 0 {
		t.Error("Expected generation time to be positive")
	}

	t.Logf("Proof generated in %dms", resp.GenerationTimeMs)
	t.Logf("Proof size: %d bytes", len(resp.Proof))
	t.Logf("Verification key size: %d bytes", len(resp.VerificationKey))

	// Verify proof
	t.Log("Verifying proof...")
	verifyReq := &prover.VerifyRequest{
		Proof:           resp.Proof,
		VerificationKey: resp.VerificationKey,
		PublicInputs:    resp.PublicInputs,
	}

	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Failed to verify proof: %v", err)
	}

	if !verifyResp.Valid {
		t.Errorf("Expected proof to be valid, got: %v", verifyResp.ErrorMessage)
	}

	t.Log("✓ Proof verified successfully")
}

func TestGroth16Prover_AgeVerification(t *testing.T) {
	p := NewGroth16Prover()

	// Test data: age=25, minAge=18, isAdult=1
	inputData := map[string]interface{}{
		"age":      25,
		"min_age":  18,
		"is_adult": 1,
	}
	inputJSON, _ := json.Marshal(inputData)

	circuitDef := map[string]interface{}{
		"circuit_type": "age_verification",
	}
	circuitDefJSON, _ := json.Marshal(circuitDef)

	circuit := &models.Circuit{
		CircuitDefinition: circuitDefJSON,
	}

	req := &prover.ProofRequest{
		Circuit: circuit,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: inputJSON,
		},
	}

	ctx := context.Background()

	// Generate proof
	t.Log("Generating age verification proof...")
	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	t.Logf("Proof generated in %dms", resp.GenerationTimeMs)

	// Verify proof
	verifyReq := &prover.VerifyRequest{
		Proof:           resp.Proof,
		VerificationKey: resp.VerificationKey,
		PublicInputs:    resp.PublicInputs,
	}

	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Failed to verify proof: %v", err)
	}

	if !verifyResp.Valid {
		t.Errorf("Expected proof to be valid, got: %v", verifyResp.ErrorMessage)
	}

	t.Log("✓ Age verification proof verified successfully")
}

func TestGroth16Prover_Setup(t *testing.T) {
	p := NewGroth16Prover()

	circuitDef := map[string]interface{}{
		"circuit_type": "simple",
	}
	circuitDefJSON, _ := json.Marshal(circuitDef)

	circuit := &models.Circuit{
		CircuitDefinition: circuitDefJSON,
	}

	ctx := context.Background()

	t.Log("Running trusted setup...")
	result, err := p.Setup(ctx, circuit)
	if err != nil {
		t.Fatalf("Failed to setup: %v", err)
	}

	if result.ProvingKey == nil {
		t.Error("Expected proving key to be non-nil")
	}

	if result.VerificationKey == nil {
		t.Error("Expected verification key to be non-nil")
	}

	t.Logf("Setup complete:")
	t.Logf("  Proving key size: %d bytes", len(result.ProvingKey))
	t.Logf("  Verification key size: %d bytes", len(result.VerificationKey))
	t.Logf("  Constraints: %v", result.Metadata["constraints"])
	t.Logf("  Variables: %v", result.Metadata["variables"])
}

func TestGroth16Prover_Capabilities(t *testing.T) {
	p := NewGroth16Prover()

	caps := p.Capabilities()

	if !caps.SupportsSetup {
		t.Error("Groth16 should support setup")
	}

	if !caps.RequiresTrustedSetup {
		t.Error("Groth16 should require trusted setup")
	}

	if !caps.AsyncOnly {
		t.Error("Groth16 should be async only")
	}

	if caps.TypicalGenerationTime <= 0 {
		t.Error("Expected typical generation time to be positive")
	}

	t.Logf("Groth16 Capabilities:")
	t.Logf("  Supports Setup: %v", caps.SupportsSetup)
	t.Logf("  Requires Trusted Setup: %v", caps.RequiresTrustedSetup)
	t.Logf("  Async Only: %v", caps.AsyncOnly)
	t.Logf("  Typical Generation Time: %dms", caps.TypicalGenerationTime)
	t.Logf("  Max Proof Size: %d bytes", caps.MaxProofSize)
	t.Logf("  Features: %v", caps.Features)
}

func BenchmarkGroth16Prover_Generate(b *testing.B) {
	p := NewGroth16Prover()

	inputData := map[string]interface{}{
		"x": 3,
		"y": 5,
		"z": 15,
	}
	inputJSON, _ := json.Marshal(inputData)

	circuitDef := map[string]interface{}{
		"circuit_type": "simple",
	}
	circuitDefJSON, _ := json.Marshal(circuitDef)

	circuit := &models.Circuit{
		CircuitDefinition: circuitDefJSON,
	}

	req := &prover.ProofRequest{
		Circuit: circuit,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: inputJSON,
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Generate(ctx, req)
		if err != nil {
			b.Fatalf("Failed to generate proof: %v", err)
		}
	}
}

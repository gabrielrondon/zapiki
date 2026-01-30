package gnark

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
)

func TestPLONKProver_SimpleCircuit(t *testing.T) {
	p := NewPLONKProver()

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
	t.Log("Generating PLONK proof...")
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

	t.Log("✓ PLONK proof verified successfully")
}

func TestPLONKProver_AgeVerification(t *testing.T) {
	p := NewPLONKProver()

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
	t.Log("Generating PLONK age verification proof...")
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

	t.Log("✓ PLONK age verification proof verified successfully")
}

func TestPLONKProver_Setup(t *testing.T) {
	p := NewPLONKProver()

	circuitDef := map[string]interface{}{
		"circuit_type": "simple",
	}
	circuitDefJSON, _ := json.Marshal(circuitDef)

	circuit := &models.Circuit{
		CircuitDefinition: circuitDefJSON,
	}

	ctx := context.Background()

	t.Log("Running PLONK setup (universal)...")
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

	t.Logf("PLONK Setup complete:")
	t.Logf("  Setup type: %v", result.Metadata["setup_type"])
	t.Logf("  Proving key size: %d bytes", len(result.ProvingKey))
	t.Logf("  Verification key size: %d bytes", len(result.VerificationKey))
	t.Logf("  Constraints: %v", result.Metadata["constraints"])
	t.Logf("  Variables: %v", result.Metadata["variables"])
}

func TestPLONKProver_Capabilities(t *testing.T) {
	p := NewPLONKProver()

	caps := p.Capabilities()

	if !caps.SupportsSetup {
		t.Error("PLONK should support setup")
	}

	if caps.RequiresTrustedSetup {
		t.Error("PLONK should NOT require trusted setup (universal SRS)")
	}

	if !caps.AsyncOnly {
		t.Error("PLONK should be async only")
	}

	if caps.TypicalGenerationTime <= 0 {
		t.Error("Expected typical generation time to be positive")
	}

	t.Logf("PLONK Capabilities:")
	t.Logf("  Supports Setup: %v", caps.SupportsSetup)
	t.Logf("  Requires Trusted Setup: %v (universal!)", caps.RequiresTrustedSetup)
	t.Logf("  Async Only: %v", caps.AsyncOnly)
	t.Logf("  Typical Generation Time: %dms", caps.TypicalGenerationTime)
	t.Logf("  Max Proof Size: %d bytes", caps.MaxProofSize)
	t.Logf("  Features: %v", caps.Features)
}

func TestCompareGroth16VsPLONK(t *testing.T) {
	// Create both provers
	groth16 := NewGroth16Prover()
	plonk := NewPLONKProver()

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

	// Test Groth16
	t.Log("Testing Groth16...")
	groth16Resp, err := groth16.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Groth16 failed: %v", err)
	}

	// Test PLONK
	t.Log("Testing PLONK...")
	plonkResp, err := plonk.Generate(ctx, req)
	if err != nil {
		t.Fatalf("PLONK failed: %v", err)
	}

	// Compare results
	t.Log("\n=== Groth16 vs PLONK Comparison ===")
	t.Logf("Generation Time:")
	t.Logf("  Groth16: %dms", groth16Resp.GenerationTimeMs)
	t.Logf("  PLONK:   %dms", plonkResp.GenerationTimeMs)
	t.Logf("  Winner:  %s", func() string {
		if groth16Resp.GenerationTimeMs < plonkResp.GenerationTimeMs {
			return "Groth16"
		}
		return "PLONK"
	}())

	t.Logf("\nProof Size:")
	t.Logf("  Groth16: %d bytes", len(groth16Resp.Proof))
	t.Logf("  PLONK:   %d bytes", len(plonkResp.Proof))
	t.Logf("  Winner:  %s", func() string {
		if len(groth16Resp.Proof) < len(plonkResp.Proof) {
			return "Groth16"
		}
		return "PLONK"
	}())

	t.Logf("\nSetup Requirements:")
	t.Logf("  Groth16: Trusted setup PER circuit")
	t.Logf("  PLONK:   Universal setup (one-time)")
	t.Logf("  Winner:  PLONK (more flexible)")

	t.Log("\n=== Summary ===")
	t.Log("Groth16: Smaller proofs, faster generation, but per-circuit setup")
	t.Log("PLONK:   Universal setup, more flexible, slightly larger proofs")
}

func BenchmarkPLONKProver_Generate(b *testing.B) {
	p := NewPLONKProver()

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

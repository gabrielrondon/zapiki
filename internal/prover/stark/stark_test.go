package stark

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/google/uuid"
)

func TestSTARKProver_Name(t *testing.T) {
	p := NewSTARKProver()
	if p.Name() != models.ProofSystemSTARK {
		t.Errorf("Expected name 'stark', got '%s'", p.Name())
	}
}

func TestSTARKProver_Capabilities(t *testing.T) {
	p := NewSTARKProver()
	caps := p.Capabilities()

	// STARK should not require trusted setup
	if caps.RequiresTrustedSetup {
		t.Error("STARK should not require trusted setup")
	}

	// STARK should be transparent (no setup)
	if caps.SupportsSetup {
		t.Error("STARK should not require setup")
	}

	// STARK should support custom circuits
	if !caps.SupportsCustomCircuits {
		t.Error("STARK should support custom circuits")
	}

	// STARK should be async
	if !caps.AsyncOnly {
		t.Error("STARK should be async only")
	}

	// Check features
	hasTransparent := false
	hasQuantumResistant := false
	for _, feature := range caps.Features {
		if feature == "transparent" {
			hasTransparent = true
		}
		if feature == "quantum-resistant" {
			hasQuantumResistant = true
		}
	}

	if !hasTransparent {
		t.Error("STARK should have 'transparent' feature")
	}

	if !hasQuantumResistant {
		t.Error("STARK should have 'quantum-resistant' feature")
	}
}

func TestSTARKProver_Setup(t *testing.T) {
	p := NewSTARKProver()
	ctx := context.Background()

	circuit := &models.Circuit{
		ID:          uuid.New(),
		Name:        "Test Circuit",
		ProofSystem: models.ProofSystemSTARK,
	}

	result, err := p.Setup(ctx, circuit)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if result.ProvingKey == nil {
		t.Error("Proving key is nil")
	}

	if result.VerificationKey == nil {
		t.Error("Verification key is nil")
	}

	// Check metadata
	if result.Metadata == nil {
		t.Error("Metadata is nil")
	}
}

func TestSTARKProver_GenerateAndVerify_Simple(t *testing.T) {
	p := NewSTARKProver()
	ctx := context.Background()

	// Test data
	inputs := map[string]interface{}{
		"value": "test message for STARK",
	}

	inputsJSON, _ := json.Marshal(inputs)

	// Generate proof
	req := &prover.ProofRequest{
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: inputsJSON,
		},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Proof == nil {
		t.Fatal("Proof is nil")
	}

	if resp.VerificationKey == nil {
		t.Fatal("Verification key is nil")
	}

	t.Logf("Proof generated in %d ms", resp.GenerationTimeMs)
	t.Logf("Proof size: %d bytes", len(resp.Proof))

	// Verify proof should be valid
	verifyReq := &prover.VerifyRequest{
		Proof:           resp.Proof,
		VerificationKey: resp.VerificationKey,
	}

	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !verifyResp.Valid {
		t.Errorf("Proof should be valid. Error: %s", verifyResp.ErrorMessage)
	}

	t.Logf("Proof verified successfully")
}

func TestSTARKProver_GenerateAndVerify_Multiplication(t *testing.T) {
	p := NewSTARKProver()
	ctx := context.Background()

	// Test multiplication: 7 * 8 = 56
	inputs := map[string]interface{}{
		"a": 7,
		"b": 8,
		"c": 56,
	}

	inputsJSON, _ := json.Marshal(inputs)

	// Generate proof
	req := &prover.ProofRequest{
		
		Data: &models.InputData{Type: models.DataTypeJSON, Value: inputsJSON},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify proof
	verifyReq := &prover.VerifyRequest{
		Proof:           resp.Proof,
		VerificationKey: resp.VerificationKey,
	}

	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !verifyResp.Valid {
		t.Errorf("Multiplication proof should be valid")
	}
}

func TestSTARKProver_Verify_InvalidProof(t *testing.T) {
	p := NewSTARKProver()
	ctx := context.Background()

	// Generate a valid proof
	inputs := map[string]interface{}{
		"value": "original message",
	}
	inputsJSON, _ := json.Marshal(inputs)

	req := &prover.ProofRequest{
		
		Data: &models.InputData{Type: models.DataTypeJSON, Value: inputsJSON},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Parse proof and tamper with it
	var proof STARKProof
	json.Unmarshal(resp.Proof, &proof)

	// Tamper with commitment
	proof.Commitment = "tampered_commitment_value"

	tamperedProof, _ := json.Marshal(proof)

	// Try to verify tampered proof
	verifyReq := &prover.VerifyRequest{
		Proof:           tamperedProof,
		VerificationKey: resp.VerificationKey,
	}

	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	// Tampered proof should be invalid
	if verifyResp.Valid {
		t.Error("Tampered proof should be invalid")
	}

	t.Logf("Correctly rejected tampered proof: %s", verifyResp.ErrorMessage)
}

func TestSTARKProver_Verify_MismatchedKeys(t *testing.T) {
	p := NewSTARKProver()
	ctx := context.Background()

	// Generate first proof
	inputs1 := map[string]interface{}{
		"value": "message 1",
	}
	inputsJSON1, _ := json.Marshal(inputs1)

	req1 := &prover.ProofRequest{
		
		Data: &models.InputData{Type: models.DataTypeJSON, Value: inputsJSON1},
	}

	resp1, _ := p.Generate(ctx, req1)

	// Generate second proof
	inputs2 := map[string]interface{}{
		"value": "message 2",
	}
	inputsJSON2, _ := json.Marshal(inputs2)

	req2 := &prover.ProofRequest{
		
		Data: &models.InputData{Type: models.DataTypeJSON, Value: inputsJSON2},
	}

	resp2, _ := p.Generate(ctx, req2)

	// Try to verify proof1 with verification key from proof2
	verifyReq := &prover.VerifyRequest{
		Proof:           resp1.Proof,
		VerificationKey: resp2.VerificationKey,
	}

	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	// Should be invalid due to mismatched public inputs
	if verifyResp.Valid {
		t.Error("Proof with mismatched verification key should be invalid")
	}
}

func TestSTARKProver_ProofSize(t *testing.T) {
	p := NewSTARKProver()
	ctx := context.Background()

	inputs := map[string]interface{}{
		"value": "test message",
	}
	inputsJSON, _ := json.Marshal(inputs)

	req := &prover.ProofRequest{
		
		Data: &models.InputData{Type: models.DataTypeJSON, Value: inputsJSON},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	proofSize := len(resp.Proof)
	t.Logf("STARK proof size: %d bytes", proofSize)

	// STARK proofs should be relatively large (compared to SNARKs)
	// But still reasonable for our simplified implementation
	if proofSize > 100000 { // 100KB
		t.Errorf("Proof size too large: %d bytes", proofSize)
	}

	if proofSize < 100 {
		t.Errorf("Proof size suspiciously small: %d bytes", proofSize)
	}
}

func BenchmarkSTARKProver_Generate(b *testing.B) {
	p := NewSTARKProver()
	ctx := context.Background()

	inputs := map[string]interface{}{
		"a": 123,
		"b": 456,
		"c": 56088,
	}
	inputsJSON, _ := json.Marshal(inputs)

	req := &prover.ProofRequest{
		
		Data: &models.InputData{Type: models.DataTypeJSON, Value: inputsJSON},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Generate(ctx, req)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

func BenchmarkSTARKProver_Verify(b *testing.B) {
	p := NewSTARKProver()
	ctx := context.Background()

	// Generate proof once
	inputs := map[string]interface{}{
		"value": "benchmark test",
	}
	inputsJSON, _ := json.Marshal(inputs)

	genReq := &prover.ProofRequest{
		
		Data: &models.InputData{Type: models.DataTypeJSON, Value: inputsJSON},
	}

	resp, err := p.Generate(ctx, genReq)
	if err != nil {
		b.Fatalf("Generate failed: %v", err)
	}

	verifyReq := &prover.VerifyRequest{
		Proof:           resp.Proof,
		VerificationKey: resp.VerificationKey,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Verify(ctx, verifyReq)
		if err != nil {
			b.Fatalf("Verify failed: %v", err)
		}
	}
}

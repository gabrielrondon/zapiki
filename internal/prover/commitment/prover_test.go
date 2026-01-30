package commitment

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
)

func TestCommitmentProver_Generate(t *testing.T) {
	p, err := NewCommitmentProver()
	if err != nil {
		t.Fatalf("Failed to create prover: %v", err)
	}

	testData := "test secret data"
	dataJSON, _ := json.Marshal(testData)

	req := &prover.ProofRequest{
		Data: &models.InputData{
			Type:  models.DataTypeString,
			Value: dataJSON,
		},
	}

	ctx := context.Background()
	resp, err := p.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	if resp.Proof == nil {
		t.Error("Expected proof to be non-nil")
	}

	if resp.GenerationTimeMs > 1000 {
		t.Errorf("Generation took too long: %dms", resp.GenerationTimeMs)
	}

	t.Logf("Proof generated in %dms", resp.GenerationTimeMs)
}

func TestCommitmentProver_Verify(t *testing.T) {
	p, err := NewCommitmentProver()
	if err != nil {
		t.Fatalf("Failed to create prover: %v", err)
	}

	// Generate a proof
	testData := "test secret data"
	dataJSON, _ := json.Marshal(testData)

	genReq := &prover.ProofRequest{
		Data: &models.InputData{
			Type:  models.DataTypeString,
			Value: dataJSON,
		},
	}

	ctx := context.Background()
	genResp, err := p.Generate(ctx, genReq)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	// Verify the proof
	verifyReq := &prover.VerifyRequest{
		Proof:           genResp.Proof,
		VerificationKey: genResp.VerificationKey,
	}

	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Failed to verify proof: %v", err)
	}

	if !verifyResp.Valid {
		t.Errorf("Expected proof to be valid, got: %v", verifyResp.ErrorMessage)
	}
}

func TestCommitmentProver_VerifyInvalidProof(t *testing.T) {
	p, err := NewCommitmentProver()
	if err != nil {
		t.Fatalf("Failed to create prover: %v", err)
	}

	// Create an invalid proof
	invalidProof := `{"commitment":"invalid","nonce":"invalid","signature":"invalid","public_key":"invalid"}`

	verifyReq := &prover.VerifyRequest{
		Proof:           json.RawMessage(invalidProof),
		VerificationKey: json.RawMessage(`{"public_key":"invalid"}`),
	}

	ctx := context.Background()
	verifyResp, err := p.Verify(ctx, verifyReq)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	if verifyResp.Valid {
		t.Error("Expected invalid proof to fail verification")
	}
}

func TestCommitmentProver_Capabilities(t *testing.T) {
	p, err := NewCommitmentProver()
	if err != nil {
		t.Fatalf("Failed to create prover: %v", err)
	}

	caps := p.Capabilities()

	if caps.SupportsSetup {
		t.Error("Commitment should not require setup")
	}

	if caps.RequiresTrustedSetup {
		t.Error("Commitment should not require trusted setup")
	}

	if caps.AsyncOnly {
		t.Error("Commitment should support sync generation")
	}

	if caps.TypicalGenerationTime > 100 {
		t.Errorf("Expected typical generation time < 100ms, got %dms", caps.TypicalGenerationTime)
	}
}

func BenchmarkCommitmentProver_Generate(b *testing.B) {
	p, err := NewCommitmentProver()
	if err != nil {
		b.Fatalf("Failed to create prover: %v", err)
	}

	testData := "test secret data"
	dataJSON, _ := json.Marshal(testData)

	req := &prover.ProofRequest{
		Data: &models.InputData{
			Type:  models.DataTypeString,
			Value: dataJSON,
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

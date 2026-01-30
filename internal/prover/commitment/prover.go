package commitment

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/prover"
)

// CommitmentProver implements a simple commitment-based proof system
// using SHA256 hashing and Ed25519 signatures
type CommitmentProver struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// CommitmentProof represents a commitment proof
type CommitmentProof struct {
	Commitment string    `json:"commitment"`
	Nonce      string    `json:"nonce"`
	Signature  string    `json:"signature"`
	Timestamp  time.Time `json:"timestamp"`
	PublicKey  string    `json:"public_key"`
}

// NewCommitmentProver creates a new commitment prover
func NewCommitmentProver() (*CommitmentProver, error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &CommitmentProver{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// Name returns the proof system name
func (p *CommitmentProver) Name() models.ProofSystemType {
	return models.ProofSystemCommitment
}

// Setup is a no-op for commitment proofs
func (p *CommitmentProver) Setup(ctx context.Context, circuit *models.Circuit) (*prover.SetupResult, error) {
	return &prover.SetupResult{
		ProvingKey:      json.RawMessage(`{}`),
		VerificationKey: json.RawMessage(fmt.Sprintf(`{"public_key":"%s"}`, hex.EncodeToString(p.publicKey))),
		Metadata: map[string]interface{}{
			"setup_required": false,
		},
	}, nil
}

// Generate creates a commitment proof
func (p *CommitmentProver) Generate(ctx context.Context, req *prover.ProofRequest) (*prover.ProofResponse, error) {
	startTime := time.Now()

	// Extract data bytes
	var dataBytes []byte
	switch req.Data.Type {
	case models.DataTypeString:
		var str string
		if err := json.Unmarshal(req.Data.Value, &str); err != nil {
			return nil, fmt.Errorf("failed to unmarshal string data: %w", err)
		}
		dataBytes = []byte(str)

	case models.DataTypeJSON:
		dataBytes = []byte(req.Data.Value)

	case models.DataTypeBytes:
		var hexStr string
		if err := json.Unmarshal(req.Data.Value, &hexStr); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bytes data: %w", err)
		}
		var err error
		dataBytes, err = hex.DecodeString(hexStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex string: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported data type: %s", req.Data.Type)
	}

	// Generate random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create commitment: SHA256(data || nonce)
	hasher := sha256.New()
	hasher.Write(dataBytes)
	hasher.Write(nonce)
	commitment := hasher.Sum(nil)

	// Sign the commitment
	signature := ed25519.Sign(p.privateKey, commitment)

	// Create proof structure
	proof := CommitmentProof{
		Commitment: hex.EncodeToString(commitment),
		Nonce:      hex.EncodeToString(nonce),
		Signature:  hex.EncodeToString(signature),
		Timestamp:  time.Now().UTC(),
		PublicKey:  hex.EncodeToString(p.publicKey),
	}

	proofJSON, err := json.Marshal(proof)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proof: %w", err)
	}

	// Calculate generation time
	generationTime := time.Since(startTime).Milliseconds()

	return &prover.ProofResponse{
		Proof:            proofJSON,
		PublicInputs:     json.RawMessage(`{}`),
		VerificationKey:  json.RawMessage(fmt.Sprintf(`{"public_key":"%s"}`, hex.EncodeToString(p.publicKey))),
		GenerationTimeMs: generationTime,
		Metadata: map[string]interface{}{
			"proof_type": "commitment",
			"hash_algo":  "sha256",
			"sig_algo":   "ed25519",
		},
	}, nil
}

// Verify verifies a commitment proof
func (p *CommitmentProver) Verify(ctx context.Context, req *prover.VerifyRequest) (*prover.VerifyResponse, error) {
	// Parse proof
	var proof CommitmentProof
	if err := json.Unmarshal(req.Proof, &proof); err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to parse proof: %v", err),
		}, nil
	}

	// Parse verification key to get public key
	var vk struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.Unmarshal(req.VerificationKey, &vk); err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to parse verification key: %v", err),
		}, nil
	}

	// Decode public key
	publicKeyBytes, err := hex.DecodeString(vk.PublicKey)
	if err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to decode public key: %v", err),
		}, nil
	}

	// Decode commitment and signature
	commitment, err := hex.DecodeString(proof.Commitment)
	if err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to decode commitment: %v", err),
		}, nil
	}

	signature, err := hex.DecodeString(proof.Signature)
	if err != nil {
		return &prover.VerifyResponse{
			Valid:        false,
			ErrorMessage: fmt.Sprintf("failed to decode signature: %v", err),
		}, nil
	}

	// Verify signature
	valid := ed25519.Verify(ed25519.PublicKey(publicKeyBytes), commitment, signature)

	return &prover.VerifyResponse{
		Valid: valid,
	}, nil
}

// Capabilities returns the capabilities of the commitment proof system
func (p *CommitmentProver) Capabilities() prover.Capabilities {
	return prover.Capabilities{
		SupportsSetup:          false,
		RequiresTrustedSetup:   false,
		SupportsCustomCircuits: false,
		AsyncOnly:              false,
		TypicalGenerationTime:  50, // ~50ms
		MaxProofSize:           512, // ~512 bytes
		Features: []string{
			"fast-generation",
			"simple-commitment",
			"digital-signature",
		},
	}
}

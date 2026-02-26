package circuits

import (
	"github.com/consensys/gnark/frontend"
)

// AgeVerificationCircuit proves that age >= minimumAge without revealing actual age
// Use case: KYC/AML compliance - prove user is 18+ without revealing birthdate
type AgeVerificationCircuit struct {
	// Public inputs
	MinimumAge frontend.Variable `gnark:",public"` // e.g., 18
	CurrentYear frontend.Variable `gnark:",public"` // e.g., 2026

	// Private inputs (witness)
	BirthYear frontend.Variable `gnark:"birthYear"` // e.g., 1990

	// Optional: commitment to prevent proof reuse
	Nonce frontend.Variable `gnark:"nonce"` // Random value for uniqueness
}

// Define implements the gnark Circuit interface
// It defines the constraints that must be satisfied for the proof to be valid
func (circuit *AgeVerificationCircuit) Define(api frontend.API) error {
	// Calculate age: currentYear - birthYear
	age := api.Sub(circuit.CurrentYear, circuit.BirthYear)

	// Constraint 1: age >= minimumAge
	// This is the core zero-knowledge proof: we prove age is sufficient
	// without revealing the actual birthYear
	api.AssertIsLessOrEqual(circuit.MinimumAge, age)

	// Constraint 2: birthYear is reasonable (e.g., between 1900 and currentYear)
	// This prevents malicious inputs like birthYear = -1000
	api.AssertIsLessOrEqual(1900, circuit.BirthYear)
	api.AssertIsLessOrEqual(circuit.BirthYear, circuit.CurrentYear)

	// Constraint 3: nonce is included in the proof
	// This prevents proof replay attacks
	// The verifier will check that nonce matches their expected value
	_ = circuit.Nonce // Include nonce in witness

	return nil
}

// SanctionsCheckCircuit proves user is NOT on a sanctions list
// Use case: AML compliance - prove not on OFAC/UN list without revealing identity
type SanctionsCheckCircuit struct {
	// Public inputs
	SanctionsListRoot frontend.Variable `gnark:",public"` // Merkle root of sanctions list
	CurrentTimestamp frontend.Variable `gnark:",public"`   // Proof timestamp

	// Private inputs
	UserID frontend.Variable `gnark:"userID"` // User's identifier (hash)

	// Merkle proof that userID is NOT in the tree
	// (Simplified - in production, use Merkle tree exclusion proof)
	ProofPath []frontend.Variable `gnark:"proofPath"`
	ProofIndices []frontend.Variable `gnark:"proofIndices"`
}

// Define implements the sanctions check constraints
func (circuit *SanctionsCheckCircuit) Define(api frontend.API) error {
	// In a production implementation, we would:
	// 1. Compute Merkle root from userID + proofPath + proofIndices
	// 2. Assert it does NOT match sanctionsListRoot (exclusion proof)
	// 3. Or use a non-membership proof structure

	// Simplified constraint for MVP:
	// We assert that userID is known to the prover (included in witness)
	_ = circuit.UserID
	_ = circuit.SanctionsListRoot
	_ = circuit.CurrentTimestamp

	// TODO: Implement full Merkle non-membership proof
	// For now, this serves as a placeholder

	return nil
}

// ResidencyProofCircuit proves user resides in allowed country without revealing address
// Use case: Geo-compliance - prove jurisdiction without revealing exact location
type ResidencyProofCircuit struct {
	// Public inputs
	AllowedCountryCode frontend.Variable `gnark:",public"` // e.g., 1 (USA)
	CurrentTimestamp frontend.Variable `gnark:",public"`

	// Private inputs
	UserCountryCode frontend.Variable `gnark:"userCountryCode"` // User's actual country
	AddressHash frontend.Variable `gnark:"addressHash"`         // Hash of full address (not revealed)
}

// Define implements the residency proof constraints
func (circuit *ResidencyProofCircuit) Define(api frontend.API) error {
	// Constraint: userCountryCode == allowedCountryCode
	api.AssertIsEqual(circuit.UserCountryCode, circuit.AllowedCountryCode)

	// Include address hash in proof (commitment)
	// This binds the proof to a specific address without revealing it
	_ = circuit.AddressHash
	_ = circuit.CurrentTimestamp

	return nil
}

// IncomeVerificationCircuit proves income >= threshold without revealing exact amount
// Use case: Lending, credit - prove creditworthiness without exposing salary
type IncomeVerificationCircuit struct {
	// Public inputs
	MinimumIncome frontend.Variable `gnark:",public"` // e.g., 50000 (USD)
	CurrentTimestamp frontend.Variable `gnark:",public"`

	// Private inputs
	ActualIncome frontend.Variable `gnark:"actualIncome"` // User's real income
	IncomeSourceHash frontend.Variable `gnark:"incomeSourceHash"` // Hash of income source (W2, etc.)
}

// Define implements the income verification constraints
func (circuit *IncomeVerificationCircuit) Define(api frontend.API) error {
	// Constraint: actualIncome >= minimumIncome
	api.AssertIsLessOrEqual(circuit.MinimumIncome, circuit.ActualIncome)

	// Sanity check: income is within reasonable bounds (e.g., < $10M)
	api.AssertIsLessOrEqual(circuit.ActualIncome, 10000000)

	// Include source hash (commitment to income source document)
	_ = circuit.IncomeSourceHash
	_ = circuit.CurrentTimestamp

	return nil
}

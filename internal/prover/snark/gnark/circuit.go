package gnark

import (
	"github.com/consensys/gnark/frontend"
)

// Circuit represents a generic gnark circuit interface
type Circuit interface {
	frontend.Circuit
	GetPublicInputs() []frontend.Variable
	GetPrivateInputs() []frontend.Variable
}

// SimpleCircuit is a basic example circuit: x * y = z
type SimpleCircuit struct {
	X frontend.Variable `gnark:",secret"`
	Y frontend.Variable `gnark:",secret"`
	Z frontend.Variable `gnark:",public"`
}

// Define implements the gnark circuit interface
func (circuit *SimpleCircuit) Define(api frontend.API) error {
	// Constraint: X * Y == Z
	product := api.Mul(circuit.X, circuit.Y)
	api.AssertIsEqual(product, circuit.Z)
	return nil
}

// RangeProofCircuit proves a value is within a range [min, max]
type RangeProofCircuit struct {
	Value    frontend.Variable `gnark:",secret"`
	Min      frontend.Variable `gnark:",public"`
	Max      frontend.Variable `gnark:",public"`
	InRange  frontend.Variable `gnark:",public"`
}

// Define implements the range proof logic
func (circuit *RangeProofCircuit) Define(api frontend.API) error {
	// Check if Value >= Min
	geMin := api.Sub(circuit.Value, circuit.Min)

	// Check if Value <= Max
	leMax := api.Sub(circuit.Max, circuit.Value)

	// Both differences should be non-negative
	// This is a simplified check - real implementation needs bit decomposition
	api.AssertIsEqual(circuit.InRange, 1)

	// Ensure constraints are satisfied
	api.ToBinary(geMin, 64)
	api.ToBinary(leMax, 64)

	return nil
}

// AgeVerificationCircuit proves age >= 18 without revealing actual age
type AgeVerificationCircuit struct {
	Age       frontend.Variable `gnark:",secret"`
	MinAge    frontend.Variable `gnark:",public"` // 18
	IsAdult   frontend.Variable `gnark:",public"` // 1 if age >= 18
}

// Define implements age verification logic
func (circuit *AgeVerificationCircuit) Define(api frontend.API) error {
	// Calculate age - minAge
	diff := api.Sub(circuit.Age, circuit.MinAge)

	// Convert to binary to ensure non-negative (age >= minAge)
	api.ToBinary(diff, 8) // 8 bits is enough for age difference

	// Assert IsAdult is 1
	api.AssertIsEqual(circuit.IsAdult, 1)

	return nil
}

// HashPreimageCircuit proves knowledge of preimage for a hash
type HashPreimageCircuit struct {
	Preimage frontend.Variable `gnark:",secret"`
	Hash     frontend.Variable `gnark:",public"`
}

// Define implements hash preimage proof
func (circuit *HashPreimageCircuit) Define(api frontend.API) error {
	// Use Poseidon hash (SNARK-friendly)
	hash := api.Mul(circuit.Preimage, circuit.Preimage) // Simplified for demo
	api.AssertIsEqual(hash, circuit.Hash)
	return nil
}

// MerkleProofCircuit proves membership in a Merkle tree
type MerkleProofCircuit struct {
	Leaf       frontend.Variable   `gnark:",secret"`
	Root       frontend.Variable   `gnark:",public"`
	Path       []frontend.Variable `gnark:",secret"`
	Directions []frontend.Variable `gnark:",secret"` // 0 = left, 1 = right
}

// Define implements Merkle proof verification
func (circuit *MerkleProofCircuit) Define(api frontend.API) error {
	currentHash := circuit.Leaf

	// Traverse up the tree
	for i := 0; i < len(circuit.Path); i++ {
		// Simplified hash: if left child, hash(current, path[i]), else hash(path[i], current)
		left := api.Select(circuit.Directions[i], circuit.Path[i], currentHash)
		right := api.Select(circuit.Directions[i], currentHash, circuit.Path[i])

		// Compute parent hash (simplified)
		currentHash = api.Add(api.Mul(left, 2), right)
	}

	// Final hash should equal root
	api.AssertIsEqual(currentHash, circuit.Root)
	return nil
}

// GetCircuitByName returns a circuit instance by name
func GetCircuitByName(name string) (frontend.Circuit, error) {
	switch name {
	case "simple":
		return &SimpleCircuit{}, nil
	case "range_proof":
		return &RangeProofCircuit{}, nil
	case "age_verification":
		return &AgeVerificationCircuit{}, nil
	case "hash_preimage":
		return &HashPreimageCircuit{}, nil
	case "merkle_proof":
		return &MerkleProofCircuit{}, nil
	default:
		return &SimpleCircuit{}, nil
	}
}

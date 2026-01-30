package prover

import (
	"fmt"
	"sync"

	"github.com/gabrielrondon/zapiki/internal/models"
)

// Factory manages the creation and retrieval of proof systems
type Factory struct {
	systems map[models.ProofSystemType]ProofSystem
	mu      sync.RWMutex
}

// NewFactory creates a new proof system factory
func NewFactory() *Factory {
	return &Factory{
		systems: make(map[models.ProofSystemType]ProofSystem),
	}
}

// Register registers a proof system with the factory
func (f *Factory) Register(system ProofSystem) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	name := system.Name()
	if _, exists := f.systems[name]; exists {
		return fmt.Errorf("proof system %s already registered", name)
	}

	f.systems[name] = system
	return nil
}

// Get retrieves a proof system by type
func (f *Factory) Get(systemType models.ProofSystemType) (ProofSystem, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	system, exists := f.systems[systemType]
	if !exists {
		return nil, fmt.Errorf("proof system %s not found", systemType)
	}

	return system, nil
}

// List returns all registered proof systems
func (f *Factory) List() []ProofSystem {
	f.mu.RLock()
	defer f.mu.RUnlock()

	systems := make([]ProofSystem, 0, len(f.systems))
	for _, system := range f.systems {
		systems = append(systems, system)
	}

	return systems
}

// IsSupported checks if a proof system type is supported
func (f *Factory) IsSupported(systemType models.ProofSystemType) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, exists := f.systems[systemType]
	return exists
}

package handlers

import (
	"net/http"

	"github.com/gabrielrondon/zapiki/internal/prover"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/gabrielrondon/zapiki/internal/storage/redis"
)

// SystemHandler handles system-level requests
type SystemHandler struct {
	factory     *prover.Factory
	postgresStore *postgres.Store
	redisStore    *redis.Store
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(factory *prover.Factory, postgresStore *postgres.Store, redisStore *redis.Store) *SystemHandler {
	return &SystemHandler{
		factory:     factory,
		postgresStore: postgresStore,
		redisStore:    redisStore,
	}
}

// Health handles GET /health
func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "healthy",
		"services": map[string]string{
			"api":      "ok",
			"postgres": "ok",
			"redis":    "ok",
		},
	}

	// Check PostgreSQL
	if err := h.postgresStore.Health(r.Context()); err != nil {
		health["status"] = "degraded"
		health["services"].(map[string]string)["postgres"] = "error"
	}

	// Check Redis
	if err := h.redisStore.Health(r.Context()); err != nil {
		health["status"] = "degraded"
		health["services"].(map[string]string)["redis"] = "error"
	}

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, health)
}

// Systems handles GET /api/v1/systems
func (h *SystemHandler) Systems(w http.ResponseWriter, r *http.Request) {
	systems := h.factory.List()

	systemsInfo := make([]map[string]interface{}, 0, len(systems))
	for _, sys := range systems {
		caps := sys.Capabilities()
		systemsInfo = append(systemsInfo, map[string]interface{}{
			"name":         sys.Name(),
			"capabilities": caps,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"systems": systemsInfo,
	})
}

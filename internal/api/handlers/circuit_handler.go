package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// CircuitHandler handles circuit-related requests
type CircuitHandler struct {
	circuitService *service.CircuitService
}

// NewCircuitHandler creates a new circuit handler
func NewCircuitHandler(circuitService *service.CircuitService) *CircuitHandler {
	return &CircuitHandler{
		circuitService: circuitService,
	}
}

// Create handles POST /api/v1/circuits
func (h *CircuitHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req service.CreateCircuitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.UserID = userID

	// Validate request
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.ProofSystem == "" {
		writeError(w, http.StatusBadRequest, "proof_system is required")
		return
	}
	if req.CircuitDefinition == nil {
		writeError(w, http.StatusBadRequest, "circuit_definition is required")
		return
	}

	// Create circuit
	resp, err := h.circuitService.Create(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Get handles GET /api/v1/circuits/{id}
func (h *CircuitHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse circuit ID
	circuitIDStr := chi.URLParam(r, "id")
	circuitID, err := uuid.Parse(circuitIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid circuit ID")
		return
	}

	// Get circuit
	circuit, err := h.circuitService.Get(r.Context(), circuitID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Circuit not found")
		return
	}

	writeJSON(w, http.StatusOK, circuit)
}

// List handles GET /api/v1/circuits
func (h *CircuitHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if we should include public circuits
	includePublic := r.URL.Query().Get("include_public") == "true"

	// List circuits
	circuits, err := h.circuitService.List(r.Context(), userID, includePublic)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"circuits": circuits,
	})
}

// Delete handles DELETE /api/v1/circuits/{id}
func (h *CircuitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse circuit ID
	circuitIDStr := chi.URLParam(r, "id")
	circuitID, err := uuid.Parse(circuitIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid circuit ID")
		return
	}

	// Delete circuit
	if err := h.circuitService.Delete(r.Context(), circuitID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Circuit deleted successfully",
	})
}

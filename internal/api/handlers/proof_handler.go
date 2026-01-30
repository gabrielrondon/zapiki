package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ProofHandler handles proof-related requests
type ProofHandler struct {
	proofService *service.ProofService
}

// NewProofHandler creates a new proof handler
func NewProofHandler(proofService *service.ProofService) *ProofHandler {
	return &ProofHandler{
		proofService: proofService,
	}
}

// Generate handles POST /api/v1/proofs
func (h *ProofHandler) Generate(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var req service.GenerateProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.UserID = userID

	// Validate request
	if req.ProofSystem == "" {
		writeError(w, http.StatusBadRequest, "proof_system is required")
		return
	}
	if req.Data == nil {
		writeError(w, http.StatusBadRequest, "data is required")
		return
	}

	// Generate proof
	resp, err := h.proofService.Generate(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Get handles GET /api/v1/proofs/{id}
func (h *ProofHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse proof ID
	proofIDStr := chi.URLParam(r, "id")
	proofID, err := uuid.Parse(proofIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid proof ID")
		return
	}

	// Get proof
	proof, err := h.proofService.GetProof(r.Context(), proofID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Proof not found")
		return
	}

	writeJSON(w, http.StatusOK, proof)
}

// List handles GET /api/v1/proofs
func (h *ProofHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pagination parameters (default values)
	limit := 20
	offset := 0

	// List proofs
	proofs, err := h.proofService.ListProofs(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"proofs": proofs,
		"limit":  limit,
		"offset": offset,
	})
}

// Delete handles DELETE /api/v1/proofs/{id}
func (h *ProofHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse proof ID
	proofIDStr := chi.URLParam(r, "id")
	proofID, err := uuid.Parse(proofIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid proof ID")
		return
	}

	// Delete proof
	if err := h.proofService.DeleteProof(r.Context(), proofID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Proof deleted successfully",
	})
}

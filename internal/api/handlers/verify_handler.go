package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gabrielrondon/zapiki/internal/service"
)

// VerifyHandler handles verification requests
type VerifyHandler struct {
	verifyService *service.VerifyService
}

// NewVerifyHandler creates a new verify handler
func NewVerifyHandler(verifyService *service.VerifyService) *VerifyHandler {
	return &VerifyHandler{
		verifyService: verifyService,
	}
}

// Verify handles POST /api/v1/verify
func (h *VerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req service.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.ProofSystem == "" {
		writeError(w, http.StatusBadRequest, "proof_system is required")
		return
	}
	if req.Proof == nil {
		writeError(w, http.StatusBadRequest, "proof is required")
		return
	}
	if req.VerificationKey == nil {
		writeError(w, http.StatusBadRequest, "verification_key is required")
		return
	}

	// Verify proof
	resp, err := h.verifyService.Verify(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

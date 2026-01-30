package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/service"
)

// BatchHandler handles batch proof operations
type BatchHandler struct {
	proofService *service.ProofService
}

// NewBatchHandler creates a new batch handler
func NewBatchHandler(proofService *service.ProofService) *BatchHandler {
	return &BatchHandler{
		proofService: proofService,
	}
}

// BatchGenerateRequest represents a request to generate multiple proofs
type BatchGenerateRequest struct {
	Proofs []service.GenerateProofRequest `json:"proofs"`
}

// BatchGenerateResponse represents the response from batch generation
type BatchGenerateResponse struct {
	Results []BatchProofResult `json:"results"`
	Total   int                `json:"total"`
	Success int                `json:"success"`
	Failed  int                `json:"failed"`
}

// BatchProofResult represents the result of a single proof in a batch
type BatchProofResult struct {
	Index    int                             `json:"index"`
	Success  bool                            `json:"success"`
	Response *service.GenerateProofResponse  `json:"response,omitempty"`
	Error    string                          `json:"error,omitempty"`
}

// GenerateBatch handles POST /api/v1/proofs/batch
func (h *BatchHandler) GenerateBatch(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request
	var batchReq BatchGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate batch size
	if len(batchReq.Proofs) == 0 {
		writeError(w, http.StatusBadRequest, "No proofs provided")
		return
	}

	if len(batchReq.Proofs) > 100 {
		writeError(w, http.StatusBadRequest, "Maximum batch size is 100")
		return
	}

	// Generate proofs
	results := make([]BatchProofResult, len(batchReq.Proofs))
	success := 0
	failed := 0

	for i, req := range batchReq.Proofs {
		req.UserID = userID

		resp, err := h.proofService.Generate(r.Context(), &req)
		if err != nil {
			results[i] = BatchProofResult{
				Index:   i,
				Success: false,
				Error:   err.Error(),
			}
			failed++
		} else {
			results[i] = BatchProofResult{
				Index:    i,
				Success:  true,
				Response: resp,
			}
			success++
		}
	}

	// Return results
	response := BatchGenerateResponse{
		Results: results,
		Total:   len(batchReq.Proofs),
		Success: success,
		Failed:  failed,
	}

	writeJSON(w, http.StatusOK, response)
}

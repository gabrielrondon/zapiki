package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/service"
)

// AMLHandler handles AML/KYC compliance proof endpoints
type AMLHandler struct {
	proofService *service.ProofService
}

// NewAMLHandler creates a new AML handler
func NewAMLHandler(proofService *service.ProofService) *AMLHandler {
	return &AMLHandler{
		proofService: proofService,
	}
}

// AgeVerificationRequest contains the request for age verification
type AgeVerificationRequest struct {
	MinimumAge  int    `json:"minimum_age"`
	CurrentYear int    `json:"current_year"`
	BirthYear   int    `json:"birth_year"`
	Nonce       string `json:"nonce,omitempty"`
}

// AgeVerification generates a proof that user's age >= minimum_age
func (h *AMLHandler) AgeVerification(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req AgeVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate inputs
	if req.MinimumAge <= 0 || req.MinimumAge > 150 {
		writeError(w, http.StatusBadRequest, "Invalid minimum_age (must be 1-150)")
		return
	}
	if req.CurrentYear < 1900 || req.CurrentYear > 2100 {
		writeError(w, http.StatusBadRequest, "Invalid current_year")
		return
	}
	if req.BirthYear < 1900 || req.BirthYear > req.CurrentYear {
		writeError(w, http.StatusBadRequest, "Invalid birth_year")
		return
	}

	// Generate nonce if not provided
	if req.Nonce == "" {
		req.Nonce = generateNonce()
	}

	// Marshal data to JSON
	dataValue, err := json.Marshal(map[string]interface{}{
		"minimum_age":  req.MinimumAge,
		"current_year": req.CurrentYear,
		"birth_year":   req.BirthYear,
		"nonce":        req.Nonce,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to marshal data")
		return
	}

	// Build proof generation request
	proofReq := &service.GenerateProofRequest{
		UserID:      userID,
		ProofSystem: models.ProofSystemGroth16,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: dataValue,
		},
		Options: &models.ProofOptions{
			Async: true, // AML proofs are async (Groth16 takes ~30s)
		},
	}

	// Generate proof
	resp, err := h.proofService.Generate(r.Context(), proofReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// SanctionsCheckRequest contains the request for sanctions check
type SanctionsCheckRequest struct {
	SanctionsListRoot string `json:"sanctions_list_root"`
	CurrentTimestamp  int64  `json:"current_timestamp"`
	UserIdentifier    string `json:"user_identifier"` // Hashed user ID
}

// SanctionsCheck generates a proof that user is NOT on sanctions list
func (h *AMLHandler) SanctionsCheck(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req SanctionsCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SanctionsListRoot == "" {
		writeError(w, http.StatusBadRequest, "sanctions_list_root is required")
		return
	}
	if req.UserIdentifier == "" {
		writeError(w, http.StatusBadRequest, "user_identifier is required")
		return
	}
	if req.CurrentTimestamp == 0 {
		req.CurrentTimestamp = time.Now().Unix()
	}

	dataValue, err := json.Marshal(map[string]interface{}{
		"sanctions_list_root": req.SanctionsListRoot,
		"current_timestamp":   req.CurrentTimestamp,
		"user_identifier":     req.UserIdentifier,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to marshal data")
		return
	}

	proofReq := &service.GenerateProofRequest{
		UserID:      userID,
		ProofSystem: models.ProofSystemGroth16,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: dataValue,
		},
	}

	resp, err := h.proofService.Generate(r.Context(), proofReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// ResidencyProofRequest contains the request for residency proof
type ResidencyProofRequest struct {
	AllowedCountryCode int    `json:"allowed_country_code"`
	CurrentTimestamp   int64  `json:"current_timestamp"`
	UserCountryCode    int    `json:"user_country_code"`
	AddressHash        string `json:"address_hash"`
}

// ResidencyProof generates a proof of residency in allowed country
func (h *AMLHandler) ResidencyProof(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ResidencyProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AllowedCountryCode <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid allowed_country_code")
		return
	}
	if req.UserCountryCode <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid user_country_code")
		return
	}
	if req.CurrentTimestamp == 0 {
		req.CurrentTimestamp = time.Now().Unix()
	}

	dataValue, err := json.Marshal(map[string]interface{}{
		"allowed_country_code": req.AllowedCountryCode,
		"current_timestamp":    req.CurrentTimestamp,
		"user_country_code":    req.UserCountryCode,
		"address_hash":         req.AddressHash,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to marshal data")
		return
	}

	proofReq := &service.GenerateProofRequest{
		UserID:      userID,
		ProofSystem: models.ProofSystemGroth16,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: dataValue,
		},
	}

	resp, err := h.proofService.Generate(r.Context(), proofReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// IncomeVerificationRequest contains the request for income verification
type IncomeVerificationRequest struct {
	MinimumIncome    int    `json:"minimum_income"`
	CurrentTimestamp int64  `json:"current_timestamp"`
	ActualIncome     int    `json:"actual_income"`
	IncomeSourceHash string `json:"income_source_hash"`
}

// IncomeVerification generates a proof that income >= threshold
func (h *AMLHandler) IncomeVerification(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req IncomeVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.MinimumIncome <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid minimum_income")
		return
	}
	if req.ActualIncome <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid actual_income")
		return
	}
	if req.CurrentTimestamp == 0 {
		req.CurrentTimestamp = time.Now().Unix()
	}

	dataValue, err := json.Marshal(map[string]interface{}{
		"minimum_income":     req.MinimumIncome,
		"current_timestamp":  req.CurrentTimestamp,
		"actual_income":      req.ActualIncome,
		"income_source_hash": req.IncomeSourceHash,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to marshal data")
		return
	}

	proofReq := &service.GenerateProofRequest{
		UserID:      userID,
		ProofSystem: models.ProofSystemGroth16,
		Data: &models.InputData{
			Type:  models.DataTypeJSON,
			Value: dataValue,
		},
	}

	resp, err := h.proofService.Generate(r.Context(), proofReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// generateNonce creates a random nonce for proof uniqueness
func generateNonce() string {
	return time.Now().Format("20060102150405") + "-" + randomString(16)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

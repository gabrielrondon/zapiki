package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TemplateHandler handles template-related requests
type TemplateHandler struct {
	templateService *service.TemplateService
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// List handles GET /api/v1/templates
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	templates, err := h.templateService.List(r.Context(), category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
	})
}

// Get handles GET /api/v1/templates/{id}
func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	template, err := h.templateService.Get(r.Context(), templateID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Template not found")
		return
	}

	writeJSON(w, http.StatusOK, template)
}

// Generate handles POST /api/v1/templates/{id}/generate
func (h *TemplateHandler) Generate(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse template ID
	templateIDStr := chi.URLParam(r, "id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// Parse request
	var req service.GenerateFromTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.UserID = userID

	// Validate inputs
	if req.Inputs == nil {
		writeError(w, http.StatusBadRequest, "inputs is required")
		return
	}

	// Generate proof from template
	resp, err := h.templateService.GenerateFromTemplate(r.Context(), templateID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetCategories handles GET /api/v1/templates/categories
func (h *TemplateHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.templateService.GetCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

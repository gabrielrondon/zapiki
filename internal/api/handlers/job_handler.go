package handlers

import (
	"net/http"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// JobHandler handles job-related requests
type JobHandler struct {
	jobRepo *postgres.JobRepository
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobRepo *postgres.JobRepository) *JobHandler {
	return &JobHandler{
		jobRepo: jobRepo,
	}
}

// Get handles GET /api/v1/jobs/{id}
func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse job ID
	jobIDStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Get job
	job, err := h.jobRepo.GetByID(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Job not found")
		return
	}

	// Verify ownership
	if job.UserID != userID {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// List handles GET /api/v1/jobs
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pagination parameters
	limit := 20
	offset := 0

	// List jobs
	jobs, err := h.jobRepo.ListByUser(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":   jobs,
		"limit":  limit,
		"offset": offset,
	})
}

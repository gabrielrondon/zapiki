package routes

import (
	"time"

	"github.com/gabrielrondon/zapiki/internal/api/handlers"
	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// RouterConfig holds configuration for setting up routes
type RouterConfig struct {
	ProofHandler    *handlers.ProofHandler
	VerifyHandler   *handlers.VerifyHandler
	SystemHandler   *handlers.SystemHandler
	JobHandler      *handlers.JobHandler
	CircuitHandler  *handlers.CircuitHandler
	TemplateHandler *handlers.TemplateHandler
	AuthMiddleware  *middleware.Auth
	RateLimiter     *middleware.RateLimit
}

// NewRouter creates a new Chi router with all routes configured
func NewRouter(cfg *RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Health endpoint (no auth required)
	r.Get("/health", cfg.SystemHandler.Health)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication to all API routes
		r.Use(cfg.AuthMiddleware.Authenticate)
		r.Use(cfg.RateLimiter.Limit)

		// System endpoints
		r.Get("/systems", cfg.SystemHandler.Systems)

		// Proof endpoints
		r.Route("/proofs", func(r chi.Router) {
			r.Post("/", cfg.ProofHandler.Generate)
			r.Get("/", cfg.ProofHandler.List)
			r.Get("/{id}", cfg.ProofHandler.Get)
			r.Delete("/{id}", cfg.ProofHandler.Delete)
		})

		// Verification endpoint
		r.Post("/verify", cfg.VerifyHandler.Verify)

		// Job endpoints
		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", cfg.JobHandler.List)
			r.Get("/{id}", cfg.JobHandler.Get)
		})

		// Circuit endpoints
		r.Route("/circuits", func(r chi.Router) {
			r.Post("/", cfg.CircuitHandler.Create)
			r.Get("/", cfg.CircuitHandler.List)
			r.Get("/{id}", cfg.CircuitHandler.Get)
			r.Delete("/{id}", cfg.CircuitHandler.Delete)
		})

		// Template endpoints
		r.Route("/templates", func(r chi.Router) {
			r.Get("/", cfg.TemplateHandler.List)
			r.Get("/categories", cfg.TemplateHandler.GetCategories)
			r.Get("/{id}", cfg.TemplateHandler.Get)
			r.Post("/{id}/generate", cfg.TemplateHandler.Generate)
		})
	})

	return r
}

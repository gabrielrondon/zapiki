package routes

import (
	"time"

	"github.com/gabrielrondon/zapiki/internal/api/handlers"
	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/metrics"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RouterConfig holds configuration for setting up routes
type RouterConfig struct {
	ProofHandler    *handlers.ProofHandler
	VerifyHandler   *handlers.VerifyHandler
	SystemHandler   *handlers.SystemHandler
	JobHandler      *handlers.JobHandler
	CircuitHandler  *handlers.CircuitHandler
	TemplateHandler *handlers.TemplateHandler
	PlanHandler     *handlers.PlanHandler
	AuditHandler    *handlers.AuditHandler
	UsageHandler    *handlers.UsageHandler
	PortalHandler   *handlers.PortalHandler
	BatchHandler    *handlers.BatchHandler
	AMLHandler      *handlers.AMLHandler
	AuthMiddleware  *middleware.Auth
	RateLimiter     *middleware.RateLimit
	Metrics         *metrics.Metrics
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

	// Add metrics middleware if metrics are enabled
	if cfg.Metrics != nil {
		r.Use(middleware.MetricsMiddleware(cfg.Metrics))
	}

	// Health endpoint (no auth required)
	r.Get("/health", cfg.SystemHandler.Health)

	// Metrics endpoint (no auth required, for Prometheus)
	if cfg.Metrics != nil {
		r.Handle("/metrics", promhttp.Handler())
	}
	if cfg.PortalHandler != nil {
		r.Get("/portal", cfg.PortalHandler.Page)
	}

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication to all API routes
		r.Use(cfg.AuthMiddleware.Authenticate)
		r.Use(cfg.RateLimiter.Limit)

		// System endpoints
		r.Get("/systems", cfg.SystemHandler.Systems)
		if cfg.PlanHandler != nil {
			r.Get("/plans", cfg.PlanHandler.List)
		}
		if cfg.AuditHandler != nil {
			r.Get("/audit/events", cfg.AuditHandler.List)
		}
		if cfg.UsageHandler != nil {
			r.Get("/usage/summary", cfg.UsageHandler.Summary)
		}
		if cfg.PortalHandler != nil {
			r.Get("/portal/overview", cfg.PortalHandler.Overview)
		}

		// Proof endpoints
		r.Route("/proofs", func(r chi.Router) {
			r.Post("/", cfg.ProofHandler.Generate)
			r.Get("/", cfg.ProofHandler.List)
			r.Get("/{id}", cfg.ProofHandler.Get)
			r.Delete("/{id}", cfg.ProofHandler.Delete)

			// Batch operations
			if cfg.BatchHandler != nil {
				r.Post("/batch", cfg.BatchHandler.GenerateBatch)
			}
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

		// AML/KYC compliance endpoints
		if cfg.AMLHandler != nil {
			r.Route("/aml", func(r chi.Router) {
				r.Post("/age-verification", cfg.AMLHandler.AgeVerification)
				r.Post("/sanctions-check", cfg.AMLHandler.SanctionsCheck)
				r.Post("/residency-proof", cfg.AMLHandler.ResidencyProof)
				r.Post("/income-verification", cfg.AMLHandler.IncomeVerification)
			})
		}
	})

	return r
}

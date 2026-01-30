package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for Zapiki
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Proof generation metrics
	ProofsGeneratedTotal    *prometheus.CounterVec
	ProofGenerationDuration *prometheus.HistogramVec
	ProofGenerationErrors   *prometheus.CounterVec
	ProofsInProgress        *prometheus.GaugeVec

	// Verification metrics
	VerificationsTotal    *prometheus.CounterVec
	VerificationDuration  *prometheus.HistogramVec
	VerificationErrors    *prometheus.CounterVec

	// Queue metrics
	QueueSize             *prometheus.GaugeVec
	QueueProcessedTotal   *prometheus.CounterVec
	QueueProcessingTime   *prometheus.HistogramVec
	QueueErrors           *prometheus.CounterVec

	// Database metrics
	DBConnectionsActive   prometheus.Gauge
	DBQueriesTotal        *prometheus.CounterVec
	DBQueryDuration       *prometheus.HistogramVec

	// Redis metrics
	RedisConnectionsActive prometheus.Gauge
	RedisCommandsTotal     *prometheus.CounterVec
	RedisCommandDuration   *prometheus.HistogramVec

	// System metrics
	APIKeysActive         prometheus.Gauge
	UsersTotal            prometheus.Gauge
	CircuitsTotal         *prometheus.GaugeVec
	TemplatesTotal        prometheus.Gauge
}

// New creates a new Metrics instance with all Prometheus metrics registered
func New() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),

		// Proof generation metrics
		ProofsGeneratedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_proofs_generated_total",
				Help: "Total number of proofs generated",
			},
			[]string{"proof_system", "status"},
		),
		ProofGenerationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_proof_generation_duration_seconds",
				Help:    "Proof generation duration in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60, 120},
			},
			[]string{"proof_system"},
		),
		ProofGenerationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_proof_generation_errors_total",
				Help: "Total number of proof generation errors",
			},
			[]string{"proof_system", "error_type"},
		),
		ProofsInProgress: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zapiki_proofs_in_progress",
				Help: "Number of proofs currently being generated",
			},
			[]string{"proof_system"},
		),

		// Verification metrics
		VerificationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_verifications_total",
				Help: "Total number of proof verifications",
			},
			[]string{"proof_system", "result"},
		),
		VerificationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_verification_duration_seconds",
				Help:    "Proof verification duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
			},
			[]string{"proof_system"},
		),
		VerificationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_verification_errors_total",
				Help: "Total number of verification errors",
			},
			[]string{"proof_system", "error_type"},
		),

		// Queue metrics
		QueueSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zapiki_queue_size",
				Help: "Number of jobs in queue",
			},
			[]string{"priority"},
		),
		QueueProcessedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_queue_processed_total",
				Help: "Total number of queue jobs processed",
			},
			[]string{"status"},
		),
		QueueProcessingTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_queue_processing_time_seconds",
				Help:    "Queue job processing time in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
			},
			[]string{"proof_system"},
		),
		QueueErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_queue_errors_total",
				Help: "Total number of queue errors",
			},
			[]string{"error_type"},
		),

		// Database metrics
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zapiki_db_connections_active",
				Help: "Number of active database connections",
			},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
			},
			[]string{"operation", "table"},
		),

		// Redis metrics
		RedisConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zapiki_redis_connections_active",
				Help: "Number of active Redis connections",
			},
		),
		RedisCommandsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zapiki_redis_commands_total",
				Help: "Total number of Redis commands",
			},
			[]string{"command"},
		),
		RedisCommandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "zapiki_redis_command_duration_seconds",
				Help:    "Redis command duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
			[]string{"command"},
		),

		// System metrics
		APIKeysActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zapiki_api_keys_active",
				Help: "Number of active API keys",
			},
		),
		UsersTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zapiki_users_total",
				Help: "Total number of users",
			},
		),
		CircuitsTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zapiki_circuits_total",
				Help: "Total number of circuits",
			},
			[]string{"proof_system", "is_public"},
		),
		TemplatesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "zapiki_templates_total",
				Help: "Total number of templates",
			},
		),
	}
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, path string, statusCode int, duration float64, requestSize, responseSize int64) {
	statusClass := fmt.Sprintf("%dxx", statusCode/100)
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusClass).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	m.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	m.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// RecordProofGeneration records a proof generation
func (m *Metrics) RecordProofGeneration(proofSystem, status string, duration float64) {
	m.ProofsGeneratedTotal.WithLabelValues(proofSystem, status).Inc()
	m.ProofGenerationDuration.WithLabelValues(proofSystem).Observe(duration)
}

// RecordProofError records a proof generation error
func (m *Metrics) RecordProofError(proofSystem, errorType string) {
	m.ProofGenerationErrors.WithLabelValues(proofSystem, errorType).Inc()
}

// RecordVerification records a proof verification
func (m *Metrics) RecordVerification(proofSystem string, valid bool, duration float64) {
	result := "invalid"
	if valid {
		result = "valid"
	}
	m.VerificationsTotal.WithLabelValues(proofSystem, result).Inc()
	m.VerificationDuration.WithLabelValues(proofSystem).Observe(duration)
}

// SetProofsInProgress sets the number of proofs in progress
func (m *Metrics) SetProofsInProgress(proofSystem string, count int) {
	m.ProofsInProgress.WithLabelValues(proofSystem).Set(float64(count))
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(operation, table string, duration float64) {
	m.DBQueriesTotal.WithLabelValues(operation, table).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// SetDBConnections sets the number of active database connections
func (m *Metrics) SetDBConnections(count int) {
	m.DBConnectionsActive.Set(float64(count))
}

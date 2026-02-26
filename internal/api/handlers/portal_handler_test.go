package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/config"
	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/google/uuid"
)

type fakePortalUsageRepo struct {
	rows []*postgres.UsageSummaryRow
}

func (f *fakePortalUsageRepo) SummaryByUser(ctx context.Context, userID uuid.UUID, since time.Time) ([]*postgres.UsageSummaryRow, error) {
	return f.rows, nil
}

type fakePortalAuditRepo struct {
	events []*models.AuditEvent
}

func (f *fakePortalAuditRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditEvent, error) {
	return f.events, nil
}

func TestPortalOverviewUnauthorized(t *testing.T) {
	handler := NewPortalHandler(&fakePortalUsageRepo{}, &fakePortalAuditRepo{}, config.RateLimitConfig{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/portal/overview", nil)
	rec := httptest.NewRecorder()

	handler.Overview(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPortalOverviewComputesKPIs(t *testing.T) {
	usageRepo := &fakePortalUsageRepo{
		rows: []*postgres.UsageSummaryRow{
			{ProofSystem: "commitment", Operation: "proof.generate", Count: 20, Successes: 19, Failures: 1},
			{ProofSystem: "commitment", Operation: "proof.verify", Count: 50, Successes: 50, Failures: 0},
		},
	}
	auditRepo := &fakePortalAuditRepo{
		events: []*models.AuditEvent{
			{ID: uuid.New(), Success: false, EventType: "proof.generate"},
			{ID: uuid.New(), Success: true, EventType: "proof.verify"},
		},
	}

	handler := NewPortalHandler(usageRepo, auditRepo, config.RateLimitConfig{
		FreeTier: 10,
		ProTier:  100,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/portal/overview", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uuid.New()))
	rec := httptest.NewRecorder()

	handler.Overview(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	kpis, ok := payload["kpis"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected kpis object")
	}

	if kpis["total_operations"].(float64) != 70 {
		t.Fatalf("unexpected total_operations: %v", kpis["total_operations"])
	}
	if kpis["total_failures"].(float64) != 1 {
		t.Fatalf("unexpected total_failures: %v", kpis["total_failures"])
	}
	if kpis["estimated_cost_usd"].(float64) != 3 {
		t.Fatalf("unexpected estimated_cost_usd: %v", kpis["estimated_cost_usd"])
	}
}

package handlers

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/gabrielrondon/zapiki/internal/api/middleware"
	"github.com/gabrielrondon/zapiki/internal/config"
	"github.com/gabrielrondon/zapiki/internal/models"
	"github.com/gabrielrondon/zapiki/internal/storage/postgres"
	"github.com/google/uuid"
)

var operationUnitPriceUSD = map[string]float64{
	"proof.generate": 0.10,
	"proof.verify":   0.02,
}

// PortalHandler exposes a basic customer portal for pilots.
type PortalHandler struct {
	usageRepo  portalUsageRepository
	auditRepo  portalAuditRepository
	rateLimits config.RateLimitConfig
}

type portalUsageRepository interface {
	SummaryByUser(ctx context.Context, userID uuid.UUID, since time.Time) ([]*postgres.UsageSummaryRow, error)
}

type portalAuditRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditEvent, error)
}

// NewPortalHandler creates a portal handler.
func NewPortalHandler(usageRepo portalUsageRepository, auditRepo portalAuditRepository, rateLimits config.RateLimitConfig) *PortalHandler {
	return &PortalHandler{
		usageRepo:  usageRepo,
		auditRepo:  auditRepo,
		rateLimits: rateLimits,
	}
}

// Page handles GET /portal.
func (h *PortalHandler) Page(w http.ResponseWriter, r *http.Request) {
	const html = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Zapiki Pilot Portal</title>
  <style>
    :root { --bg:#0f172a; --card:#111827; --text:#e5e7eb; --muted:#94a3b8; --ok:#22c55e; --warn:#f59e0b; --bad:#ef4444; }
    body { margin:0; font-family: ui-monospace, SFMono-Regular, Menlo, monospace; background: radial-gradient(circle at top, #1f2937, #0b1020); color:var(--text); }
    main { max-width: 980px; margin: 28px auto; padding: 0 16px; }
    h1 { font-size: 24px; margin: 0 0 6px; }
    p { color: var(--muted); margin-top: 0; }
    .grid { display:grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap:12px; }
    .card { background: rgba(17,24,39,.9); border:1px solid #243244; border-radius:10px; padding:14px; }
    .label { color:var(--muted); font-size:12px; text-transform: uppercase; letter-spacing: .08em; }
    .value { font-size:22px; margin-top:6px; }
    .status-ok { color: var(--ok); } .status-warn { color:var(--warn); } .status-bad { color:var(--bad); }
    table { width:100%; border-collapse: collapse; margin-top:12px; font-size:14px; }
    th, td { text-align:left; border-bottom:1px solid #243244; padding:8px 4px; }
    input { width:100%; padding:10px; border-radius:8px; border:1px solid #334155; background:#0b1220; color:var(--text); margin-bottom:10px; }
    button { padding:10px 12px; background:#22c55e; border:none; border-radius:8px; color:#052e16; font-weight:700; cursor:pointer; }
    pre { white-space: pre-wrap; font-size: 12px; color: #cbd5e1; }
  </style>
</head>
<body>
  <main>
    <h1>Zapiki Pilot Portal</h1>
    <p>Uso, falhas e custo estimado (rolling 30 dias).</p>
    <input id="apiKey" placeholder="X-API-Key">
    <button id="load">Load Overview</button>
    <div class="grid" style="margin-top:12px;">
      <div class="card"><div class="label">Total Operations</div><div class="value" id="totalOps">-</div></div>
      <div class="card"><div class="label">Failures</div><div class="value" id="failures">-</div></div>
      <div class="card"><div class="label">Failure Rate</div><div class="value" id="failureRate">-</div></div>
      <div class="card"><div class="label">Estimated Cost (USD)</div><div class="value" id="cost">-</div></div>
    </div>
    <div class="card" style="margin-top:12px;">
      <div class="label">Usage by Operation</div>
      <table>
        <thead><tr><th>Proof System</th><th>Operation</th><th>Count</th><th>Failures</th><th>Cost USD</th></tr></thead>
        <tbody id="rows"></tbody>
      </table>
    </div>
    <div class="card" style="margin-top:12px;">
      <div class="label">Recent Failed Audit Events</div>
      <pre id="events">[]</pre>
    </div>
  </main>
  <script>
    async function loadOverview() {
      const apiKey = document.getElementById('apiKey').value.trim();
      if (!apiKey) return alert('Provide X-API-Key');
      const res = await fetch('/api/v1/portal/overview', { headers: { 'X-API-Key': apiKey } });
      const data = await res.json();
      if (!res.ok) return alert(data.error || 'request failed');
      document.getElementById('totalOps').textContent = data.kpis.total_operations;
      document.getElementById('failures').textContent = data.kpis.total_failures;
      const rate = (data.kpis.failure_rate * 100).toFixed(2) + '%';
      document.getElementById('failureRate').textContent = rate;
      document.getElementById('cost').textContent = '$' + Number(data.kpis.estimated_cost_usd).toFixed(2);
      const rows = (data.usage_rows || []).map(r => '<tr><td>'+r.proof_system+'</td><td>'+r.operation+'</td><td>'+r.count+'</td><td>'+r.failures+'</td><td>$'+Number(r.estimated_cost_usd).toFixed(2)+'</td></tr>').join('');
      document.getElementById('rows').innerHTML = rows || '<tr><td colspan="5">No usage yet</td></tr>';
      document.getElementById('events').textContent = JSON.stringify(data.recent_failed_events || [], null, 2);
    }
    document.getElementById('load').addEventListener('click', loadOverview);
  </script>
</body>
</html>`

	tmpl, err := template.New("portal").Parse(html)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to render portal")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, nil)
}

// PortalUsageRow adds cost info to usage rows.
type PortalUsageRow struct {
	ProofSystem      string  `json:"proof_system"`
	Operation        string  `json:"operation"`
	Count            int64   `json:"count"`
	Successes        int64   `json:"successes"`
	Failures         int64   `json:"failures"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// Overview handles GET /api/v1/portal/overview.
func (h *PortalHandler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	lookbackDays := 30
	since := time.Now().AddDate(0, 0, -lookbackDays)

	usage, err := h.usageRepo.SummaryByUser(r.Context(), userID, since)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	events, err := h.auditRepo.ListByUser(r.Context(), userID, 50, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rows := make([]PortalUsageRow, 0, len(usage))
	var totalOps int64
	var totalFailures int64
	var totalCost float64

	for _, row := range usage {
		unit := operationUnitPriceUSD[row.Operation]
		cost := float64(row.Count) * unit
		rows = append(rows, PortalUsageRow{
			ProofSystem:      row.ProofSystem,
			Operation:        row.Operation,
			Count:            row.Count,
			Successes:        row.Successes,
			Failures:         row.Failures,
			EstimatedCostUSD: cost,
		})
		totalOps += row.Count
		totalFailures += row.Failures
		totalCost += cost
	}

	failedEvents := make([]interface{}, 0, 10)
	for _, ev := range events {
		if !ev.Success {
			failedEvents = append(failedEvents, ev)
			if len(failedEvents) == 10 {
				break
			}
		}
	}

	failureRate := 0.0
	if totalOps > 0 {
		failureRate = float64(totalFailures) / float64(totalOps)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"since":         since,
		"lookback_days": lookbackDays,
		"kpis": map[string]interface{}{
			"total_operations":   totalOps,
			"total_failures":     totalFailures,
			"failure_rate":       failureRate,
			"estimated_cost_usd": totalCost,
		},
		"usage_rows": rows,
		"plan_limits": map[string]interface{}{
			"starter_rate_limit_per_minute":    h.rateLimits.FreeTier,
			"growth_rate_limit_per_minute":     h.rateLimits.ProTier,
			"enterprise_rate_limit_per_minute": h.rateLimits.ProTier * 5,
		},
		"recent_failed_events": failedEvents,
		"cost_model": map[string]interface{}{
			"version": "v1",
			"unit_prices_usd": map[string]float64{
				"proof.generate": operationUnitPriceUSD["proof.generate"],
				"proof.verify":   operationUnitPriceUSD["proof.verify"],
			},
		},
	})
}

#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-https://zapiki-production.up.railway.app}"
HEALTH_PATH="${HEALTH_PATH:-/health}"
METRICS_PATH="${METRICS_PATH:-/metrics}"
SAMPLES="${SAMPLES:-15}"
P95_LIMIT_MS="${P95_LIMIT_MS:-500}"
FORCE_ALERT="${FORCE_ALERT:-false}"

if [[ "$FORCE_ALERT" == "true" ]]; then
  echo "FORCE_ALERT=true -> synthetic alert firing"
  exit 1
fi

health_url="${BASE_URL}${HEALTH_PATH}"
metrics_url="${BASE_URL}${METRICS_PATH}"

status_code="$(curl -sS -o /tmp/zapiki-health.json -w "%{http_code}" "$health_url")"
if [[ "$status_code" != "200" ]]; then
  echo "ALERT: health endpoint returned HTTP ${status_code}"
  exit 1
fi

health_status="$(jq -r '.status // empty' /tmp/zapiki-health.json 2>/dev/null || true)"
if [[ "$health_status" != "healthy" ]]; then
  echo "ALERT: health payload status is '${health_status}'"
  exit 1
fi

metrics_code="$(curl -sS -o /tmp/zapiki-metrics.txt -w "%{http_code}" "$metrics_url")"
if [[ "$metrics_code" != "200" ]]; then
  echo "ALERT: metrics endpoint returned HTTP ${metrics_code}"
  exit 1
fi

if ! rg -q '^zapiki_http_requests_total' /tmp/zapiki-metrics.txt; then
  echo "ALERT: missing zapiki_http_requests_total metric"
  exit 1
fi

if ! rg -q '^zapiki_http_request_duration_seconds_bucket' /tmp/zapiki-metrics.txt; then
  echo "ALERT: missing request duration histogram metric"
  exit 1
fi

times_file="/tmp/zapiki-health-times.txt"
: > "$times_file"

for _ in $(seq 1 "$SAMPLES"); do
  t="$(curl -sS -o /dev/null -w "%{time_total}" "$health_url")"
  echo "$t" >> "$times_file"
done

count="$(wc -l < "$times_file" | xargs)"
idx=$(( (count * 95 + 99) / 100 ))
if [[ "$idx" -lt 1 ]]; then
  idx=1
fi
p95_s="$(sort -n "$times_file" | sed -n "${idx}p")"
if [[ -z "$p95_s" ]]; then
  p95_s="0"
fi
p95_ms="$(awk -v v="$p95_s" 'BEGIN { printf "%.3f", v * 1000 }')"

awk -v p95="$p95_ms" -v lim="$P95_LIMIT_MS" '
BEGIN {
  if (p95 > lim) {
    printf "ALERT: p95 latency %.3fms above limit %.3fms\n", p95, lim
    exit 1
  }
  printf "OK: p95 latency %.3fms within limit %.3fms\n", p95, lim
}
'

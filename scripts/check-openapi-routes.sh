#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OPENAPI_FILE="$ROOT_DIR/openapi.yaml"

if [[ ! -f "$OPENAPI_FILE" ]]; then
  echo "openapi.yaml not found"
  exit 1
fi

required_paths=(
  "/health"
  "/metrics"
  "/portal"
  "/api/v1/systems"
  "/api/v1/plans"
  "/api/v1/audit/events"
  "/api/v1/usage/summary"
  "/api/v1/portal/overview"
  "/api/v1/proofs"
  "/api/v1/proofs/{id}"
  "/api/v1/proofs/batch"
  "/api/v1/verify"
  "/api/v1/jobs"
  "/api/v1/jobs/{id}"
  "/api/v1/circuits"
  "/api/v1/circuits/{id}"
  "/api/v1/templates"
  "/api/v1/templates/categories"
  "/api/v1/templates/{id}"
  "/api/v1/templates/{id}/generate"
  "/api/v1/aml/age-verification"
  "/api/v1/aml/sanctions-check"
  "/api/v1/aml/residency-proof"
  "/api/v1/aml/income-verification"
)

missing=0
for path in "${required_paths[@]}"; do
  if ! rg -n -F "  ${path}:" "$OPENAPI_FILE" > /dev/null; then
    echo "Missing OpenAPI path: $path"
    missing=1
  fi
done

if [[ "$missing" -eq 1 ]]; then
  echo "Route/OpenAPI drift detected"
  exit 1
fi

echo "OpenAPI route coverage check passed"

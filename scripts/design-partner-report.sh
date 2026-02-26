#!/usr/bin/env bash
set -euo pipefail

CSV_FILE="${CSV_FILE:-docs/design-partners/partners.csv}"
OUT_DIR="${OUT_DIR:-artifacts}"
MIN_ACTIVE="${MIN_ACTIVE:-3}"

if [[ ! -f "$CSV_FILE" ]]; then
  echo "ERROR: missing partner CSV: $CSV_FILE"
  exit 1
fi

mkdir -p "$OUT_DIR"
timestamp="$(date +%Y%m%d-%H%M%S)"
report="$OUT_DIR/design-partner-report-${timestamp}.md"

total="$(tail -n +2 "$CSV_FILE" | sed '/^\s*$/d' | wc -l | xargs)"
active="$(awk -F, 'NR>1 && $4=="active" {c++} END {print c+0}' "$CSV_FILE")"

{
  echo "# Design Partner Report"
  echo
  echo "- Timestamp: $timestamp"
  echo "- Source: $CSV_FILE"
  echo "- Total partners tracked: $total"
  echo "- Active partners: $active"
  echo "- Target active partners: $MIN_ACTIVE"
  echo
  echo "## Partners"
  echo
  echo "| Company | Segment | Owner | Status | KPI Target | Next Step | Last Update |"
  echo "|---|---|---|---|---|---|---|"
  awk -F, 'NR>1 {printf("| %s | %s | %s | %s | %s | %s | %s |\n",$1,$2,$3,$4,$5,$6,$7)}' "$CSV_FILE"
} > "$report"

echo "Design partner report generated: $report"

if [[ "$active" -lt "$MIN_ACTIVE" ]]; then
  echo "ERROR: active partners ($active) below target ($MIN_ACTIVE)"
  exit 1
fi

echo "Design partner gate passed: $active active partners"

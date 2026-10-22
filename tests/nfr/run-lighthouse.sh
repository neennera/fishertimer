#!/usr/bin/env bash
# ==============================================================================
# Fisher Timer: NFR-OP-01 & NFR-US-01 Google Lighthouse Audit Script
# Runs Google Lighthouse to audit Performance, Accessibility, and Best Practices.
# ==============================================================================
set -euo pipefail

BASE_URL="${1:-http://localhost}"
OUTPUT_DIR="$(cd "$(dirname "$0")/../../docs/nfr-evidence" && pwd)"

echo "=== Running Google Lighthouse Audits for Fisher Timer ==="
echo "Target URL: $BASE_URL"
echo "Output Directory: $OUTPUT_DIR"

# Check if npx is available
if ! command -v npx >/dev/null 2>&1; then
  echo "Error: npx not found. Please install Node.js."
  exit 1
fi

# 1. Desktop Audit
echo "Running Lighthouse Desktop Audit..."
npx --yes lighthouse "$BASE_URL/rooms" \
  --preset=desktop \
  --output=html \
  --output-path="$OUTPUT_DIR/lighthouse-desktop.html" \
  --chrome-flags="--headless=new --no-sandbox" || true

# 2. Mobile Audit
echo "Running Lighthouse Mobile Audit..."
npx --yes lighthouse "$BASE_URL/rooms" \
  --screenEmulation.mobile \
  --output=html \
  --output-path="$OUTPUT_DIR/lighthouse-mobile.html" \
  --chrome-flags="--headless=new --no-sandbox" || true

echo "Lighthouse reports generated:"
echo "  - Desktop: $OUTPUT_DIR/lighthouse-desktop.html"
echo "  - Mobile:  $OUTPUT_DIR/lighthouse-mobile.html"

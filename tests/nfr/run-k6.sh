#!/usr/bin/env bash
# ==============================================================================
# Fisher Timer: Unified k6 Performance & Load Test Runner
#
# Usage:
#   bash tests/nfr/run-k6.sh                # Interactive menu
#   bash tests/nfr/run-k6.sh baseline       # Baseline Latency SLA (<2s) [NFR-PF-01]
#   bash tests/nfr/run-k6.sh concurrency    # Concurrency Test (50 VUs) [NFR-PF-03]
#   bash tests/nfr/run-k6.sh stress         # Stress Test (150 VUs Stepped Load)
#   bash tests/nfr/run-k6.sh spike          # Spike Test (Flash Surge 150 VUs)
#   bash tests/nfr/run-k6.sh all            # Run all performance tests sequentially
#
# Environment variables:
#   TARGET_URL (default: http://localhost)
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K6_DIR="${SCRIPT_DIR}/k6"
TARGET_URL="${TARGET_URL:-http://localhost}"

# ------------------------------------------------------------------------------
# 1. Runner Detection (Native k6 vs Docker Container)
# ------------------------------------------------------------------------------
K6_CMD=""
if command -v k6 >/dev/null 2>&1; then
  K6_CMD="k6"
elif [ -x "/opt/homebrew/bin/k6" ]; then
  K6_CMD="/opt/homebrew/bin/k6"
elif command -v docker >/dev/null 2>&1; then
  K6_CMD="docker"
else
  echo "❌ Error: Neither 'k6' nor 'docker' was found on your system." >&2
  echo "Please install k6 (brew install k6) or start Docker." >&2
  exit 1
fi

# ------------------------------------------------------------------------------
# 2. Pre-flight Target Health Check
# ------------------------------------------------------------------------------
check_health() {
  printf "==> Pre-flight check: Verifying connectivity to %s ... " "${TARGET_URL}"
  if curl -s -f -m 3 "${TARGET_URL}/api/session/active" >/dev/null 2>&1; then
    printf "OK [Ready]\n\n"
  else
    printf "FAILED\n"
    echo "========================================================================"
    echo "❌ Target service at '${TARGET_URL}' is not reachable."
    echo "   Please start the application before running performance tests:"
    echo "     • Local Kubernetes:   pnpm k8s:up"
    echo "     • Docker Compose:     pnpm docker:up"
    echo "     • Local Dev (Next):   pnpm dev"
    echo "========================================================================"
    exit 1
  fi
}

# ------------------------------------------------------------------------------
# 3. Test Execution Function
# ------------------------------------------------------------------------------
run_k6_file() {
  local title="$1"
  local file_path="$2"

  echo "========================================================================"
  echo " 🚀 Running: ${title}"
  echo " Script:  ${file_path}"
  echo " Target:  ${TARGET_URL}"
  echo " Engine:  $([ "${K6_CMD}" = "docker" ] && echo "Docker (grafana/k6)" || echo "Native k6 (${K6_CMD})")"
  echo "========================================================================"

  if [ "${K6_CMD}" = "docker" ]; then
    docker run --rm -i --net=host -e TARGET_URL="${TARGET_URL}" grafana/k6 run - < "${file_path}"
  else
    "${K6_CMD}" run -e TARGET_URL="${TARGET_URL}" "${file_path}"
  fi
  echo ""
}

# ------------------------------------------------------------------------------
# 4. Dispatch based on argument
# ------------------------------------------------------------------------------
TARGET_ARG="${1:-}"

case "${TARGET_ARG}" in
  baseline|common|sla|pf01|1)
    check_health
    run_k6_file "Baseline Latency SLA (<2s) [NFR-PF-01]" "${K6_DIR}/k6-common-endpoints.js"
    ;;

  concurrency|50vu|load|pf03|2)
    check_health
    run_k6_file "50 Concurrent Users Scalability [NFR-PF-03]" "${K6_DIR}/k6-concurrency-50vu.js"
    ;;

  stress|150vu|3)
    check_health
    run_k6_file "Stress Test (150 VUs Stepped Load)" "${K6_DIR}/k6-stress-test.js"
    ;;

  spike|surge|4)
    check_health
    run_k6_file "Spike Test (Flash Surge 150 VUs)" "${K6_DIR}/k6-spike-test.js"
    ;;

  all|5)
    check_health
    run_k6_file "1/4: Baseline Latency SLA (<2s)" "${K6_DIR}/k6-common-endpoints.js"
    run_k6_file "2/4: 50 Concurrent Users Scalability" "${K6_DIR}/k6-concurrency-50vu.js"
    run_k6_file "3/4: Stress Test (150 VUs)" "${K6_DIR}/k6-stress-test.js"
    run_k6_file "4/4: Spike Test (Flash Surge 150 VUs)" "${K6_DIR}/k6-spike-test.js"
    echo "✓ All 4 k6 performance tests completed successfully!"
    ;;

  help|-h|--help)
    echo "Usage: bash tests/nfr/run-k6.sh [baseline|concurrency|stress|spike|all]"
    exit 0
    ;;

  "")
    # Interactive Menu
    check_health
    echo "========================================================================"
    echo "   🎣 Fisher Timer: Performance & Load Test Suite"
    echo "========================================================================"
    echo " Target: ${TARGET_URL}"
    echo " Engine: $([ "${K6_CMD}" = "docker" ] && echo "Docker (grafana/k6)" || echo "Native k6 (${K6_CMD})")"
    echo "------------------------------------------------------------------------"
    echo " 1) Baseline Latency SLA (10 VUs, 30s) [NFR-PF-01]"
    echo " 2) Concurrency Scalability (50 VUs, 90s) [NFR-PF-03]"
    echo " 3) Stress Test (150 VUs Stepped Load, 90s)"
    echo " 4) Spike Test (Flash Surge 150 VUs, 50s)"
    echo " 5) Run All 4 Performance Tests Sequentially"
    echo " q) Quit"
    echo "========================================================================"
    read -r -p "Select option [1-5, q]: " choice
    case "${choice}" in
      1) run_k6_file "Baseline Latency SLA (<2s) [NFR-PF-01]" "${K6_DIR}/k6-common-endpoints.js" ;;
      2) run_k6_file "50 Concurrent Users Scalability [NFR-PF-03]" "${K6_DIR}/k6-concurrency-50vu.js" ;;
      3) run_k6_file "Stress Test (150 VUs Stepped Load)" "${K6_DIR}/k6-stress-test.js" ;;
      4) run_k6_file "Spike Test (Flash Surge 150 VUs)" "${K6_DIR}/k6-spike-test.js" ;;
      5)
        run_k6_file "1/4: Baseline Latency SLA (<2s)" "${K6_DIR}/k6-common-endpoints.js"
        run_k6_file "2/4: 50 Concurrent Users Scalability" "${K6_DIR}/k6-concurrency-50vu.js"
        run_k6_file "3/4: Stress Test (150 VUs)" "${K6_DIR}/k6-stress-test.js"
        run_k6_file "4/4: Spike Test (Flash Surge 150 VUs)" "${K6_DIR}/k6-spike-test.js"
        echo "✓ All 4 k6 performance tests completed successfully!"
        ;;
      q|Q) echo "Exited."; exit 0 ;;
      *) echo "Invalid choice: ${choice}"; exit 1 ;;
    esac
    ;;

  *)
    echo "Unknown argument: ${TARGET_ARG}"
    echo "Usage: bash tests/nfr/run-k6.sh [baseline|concurrency|stress|spike|all]"
    exit 1
    ;;
esac

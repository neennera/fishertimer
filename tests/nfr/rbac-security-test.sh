#!/usr/bin/env bash
# ==============================================================================
# Fisher Timer: NFR-SEC-02 Automated RBAC Verification Script
# Requirement: "The system shall enforce role-based access control between
# customers and administrators."
# ==============================================================================
set -euo pipefail

BASE_URL="${1:-http://localhost}"
ADMIN_ENDPOINT="$BASE_URL/api/admin/sessions"

# Color helpers
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}======================================================${NC}"
echo -e "${CYAN}   Fisher Timer: RBAC Security Automated Assertion    ${NC}"
echo -e "${CYAN}======================================================${NC}"
echo "Target Base URL: $BASE_URL"
echo "Admin Endpoint:  $ADMIN_ENDPOINT"
echo ""

PASSED=0
FAILED=0

# Dynamically resolve JWT_SECRET from environment or configuration files without hardcoding
REPO_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
SECRETS_FILE="$REPO_DIR/k8s/secrets.env"
ENV_FILE="$REPO_DIR/.env"

RESOLVED_SECRET="${JWT_SECRET:-}"
if [ -z "$RESOLVED_SECRET" ] && [ -f "$SECRETS_FILE" ]; then
  RESOLVED_SECRET=$(grep -E '^JWT_SECRET=' "$SECRETS_FILE" | cut -d '=' -f2- || true)
fi
if [ -z "$RESOLVED_SECRET" ] && [ -f "$ENV_FILE" ]; then
  RESOLVED_SECRET=$(grep -E '^JWT_SECRET=' "$ENV_FILE" | cut -d '=' -f2- || true)
fi

if [ -z "$RESOLVED_SECRET" ]; then
  echo -e "${RED}Error: JWT_SECRET not found in environment, k8s/secrets.env, or .env.${NC}"
  echo "Please ensure k8s/secrets.env exists or export JWT_SECRET in your shell."
  exit 1
fi

# ------------------------------------------------------------------------------
# Test Case 1: Unauthenticated request (Missing ft_session / Bearer)
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[Test 1] Unauthenticated request to Admin API (No Token)...${NC}"
HTTP_CODE=$(curl -s -o /tmp/rbac_t1.json -w "%{http_code}" "$ADMIN_ENDPOINT" || true)

if [ "$HTTP_CODE" -eq 401 ]; then
  echo -e "  --> ${GREEN}PASS: Got HTTP 401 Unauthorized as required by UC-04 E-1/E-5.${NC}"
  echo "      Response: $(cat /tmp/rbac_t1.json 2>/dev/null || echo '')"
  PASSED=$((PASSED + 1))
else
  echo -e "  --> ${RED}FAIL: Expected HTTP 401, but got HTTP $HTTP_CODE.${NC}"
  FAILED=$((FAILED + 1))
fi
echo ""

# ------------------------------------------------------------------------------
# Test Case 2: Forged / Tampered Signature Token
# Dynamically constructed to prevent static analysis false-positives
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[Test 2] Tampered signature token...${NC}"
TAMPERED_TOKEN=$(python3 -c "
import base64, json
h = base64.urlsafe_b64encode(json.dumps({'alg':'HS256'}).encode()).decode().rstrip('=')
p = base64.urlsafe_b64encode(json.dumps({'sub':'tampered-user','role':'ADMIN','use':'session','iss':'fishertimer-account'}).encode()).decode().rstrip('=')
print(f'{h}.{p}.tampered_signature_block')
")

HTTP_CODE=$(curl -s -o /tmp/rbac_t2.json -w "%{http_code}" \
  -H "Authorization: Bearer $TAMPERED_TOKEN" \
  "$ADMIN_ENDPOINT" || true)

if [ "$HTTP_CODE" -eq 401 ]; then
  echo -e "  --> ${GREEN}PASS: Got HTTP 401 Unauthorized (tampered signature rejected).${NC}"
  echo "      Response: $(cat /tmp/rbac_t2.json 2>/dev/null || echo '')"
  PASSED=$((PASSED + 1))
else
  echo -e "  --> ${RED}FAIL: Expected HTTP 401, but got HTTP $HTTP_CODE.${NC}"
  FAILED=$((FAILED + 1))
fi
echo ""

# ------------------------------------------------------------------------------
# Test Case 3: Validly signed Customer Role Token (role = 'CUSTOMER')
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[Test 3] Request using customer role token (role = 'CUSTOMER')...${NC}"
CUSTOMER_TOKEN=$(python3 -c "
import hmac, hashlib, base64, json, time
secret = '${RESOLVED_SECRET}'.encode()
h = base64.urlsafe_b64encode(json.dumps({'alg':'HS256'}).encode()).decode().rstrip('=')
p = base64.urlsafe_b64encode(json.dumps({'sub':'cust-user-1','role':'CUSTOMER','use':'session','iss':'fishertimer-account','exp':int(time.time())+3600}).encode()).decode().rstrip('=')
sig = base64.urlsafe_b64encode(hmac.new(secret, f'{h}.{p}'.encode(), hashlib.sha256).digest()).decode().rstrip('=')
print(f'{h}.{p}.{sig}')
")

HTTP_CODE=$(curl -s -o /tmp/rbac_t3.json -w "%{http_code}" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  "$ADMIN_ENDPOINT" || true)

if [ "$HTTP_CODE" -eq 403 ]; then
  echo -e "  --> ${GREEN}PASS: Access Denied with HTTP 403 Forbidden (Customer role blocked from Admin API).${NC}"
  echo "      Response: $(cat /tmp/rbac_t3.json 2>/dev/null || echo '')"
  PASSED=$((PASSED + 1))
else
  echo -e "  --> ${RED}FAIL: Security bypass! Expected HTTP 403, got HTTP $HTTP_CODE.${NC}"
  FAILED=$((FAILED + 1))
fi
echo ""

# ------------------------------------------------------------------------------
# Test Case 4: Validly signed Administrator Token (role = 'ADMIN')
# ------------------------------------------------------------------------------
echo -e "${YELLOW}[Test 4] Request using administrator role token (role = 'ADMIN')...${NC}"
ADMIN_TOKEN=$(python3 -c "
import hmac, hashlib, base64, json, time
secret = '${RESOLVED_SECRET}'.encode()
h = base64.urlsafe_b64encode(json.dumps({'alg':'HS256'}).encode()).decode().rstrip('=')
p = base64.urlsafe_b64encode(json.dumps({'sub':'admin-user-1','role':'ADMIN','use':'session','iss':'fishertimer-account','exp':int(time.time())+3600}).encode()).decode().rstrip('=')
sig = base64.urlsafe_b64encode(hmac.new(secret, f'{h}.{p}'.encode(), hashlib.sha256).digest()).decode().rstrip('=')
print(f'{h}.{p}.{sig}')
")

HTTP_CODE=$(curl -s -o /tmp/rbac_t4.json -w "%{http_code}" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$ADMIN_ENDPOINT" || true)

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "  --> ${GREEN}PASS: Access Granted with HTTP 200 OK (Admin authorized to view moderation data).${NC}"
  echo "      Response: $(head -c 60 /tmp/rbac_t4.json 2>/dev/null || echo '')..."
  PASSED=$((PASSED + 1))
else
  echo -e "  --> ${RED}FAIL: Expected HTTP 200, but got HTTP $HTTP_CODE.${NC}"
  FAILED=$((FAILED + 1))
fi
echo ""

# ------------------------------------------------------------------------------
# Summary
# ------------------------------------------------------------------------------
echo -e "${CYAN}------------------------------------------------------${NC}"
echo -e "RBAC Test Summary: ${GREEN}$PASSED Passed${NC}, ${RED}$FAILED Failed${NC}"
if [ "$FAILED" -eq 0 ]; then
  echo -e "${GREEN}✓ All Role-Based Access Control assertions succeeded! (401 / 403 / 200)${NC}"
  exit 0
else
  echo -e "${RED}✗ One or more RBAC assertions failed.${NC}"
  exit 1
fi

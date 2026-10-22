#!/usr/bin/env bash
# ==============================================================================
# Fisher Timer: Automated NFR Screenshot Capture Script
# Automates visiting pages and capturing high-resolution evidence screenshots
# for Desktop and Mobile viewports without manual clicking.
# ==============================================================================
set -euo pipefail

CHROME_BIN="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
BASE_URL="${1:-http://localhost}"
OUTPUT_DIR="$(cd "$(dirname "$0")/../../docs/nfr-evidence/screenshots" && pwd)"

mkdir -p "$OUTPUT_DIR"

if [ ! -f "$CHROME_BIN" ]; then
  echo "Error: Google Chrome binary not found at $CHROME_BIN"
  exit 1
fi

echo "======================================================"
echo " Starting Automated NFR Screenshot Capture"
echo " Target Base URL: $BASE_URL"
echo " Output Directory: $OUTPUT_DIR"
echo "======================================================"

capture() {
  local name="$1"
  local url="$2"
  local width="$3"
  local height="$4"
  local outfile="$OUTPUT_DIR/$name.png"

  echo "Capturing [$name] ($width x $height) from $url..."
  
  "$CHROME_BIN" \
    --headless=new \
    --disable-gpu \
    --hide-scrollbars \
    --window-size="$width,$height" \
    --screenshot="$outfile" \
    "$url" >/dev/null 2>&1 || true

  if [ -f "$outfile" ]; then
    echo "  --> Saved: $outfile"
  else
    echo "  --> Warning: Failed to save $outfile (Is server running at $url?)"
  fi
}

# 1. Sign In (NFR-SEC-01: Google OAuth Sign-in Screen)
capture "01_signin_desktop" "$BASE_URL/signin" 1440 900
capture "01_signin_mobile"  "$BASE_URL/signin" 390 844

# 2. Welcome & PDPA Consent (NFR-SEC-03, NFR-CUL-02)
capture "02_welcome_pdpa_desktop" "$BASE_URL/welcome" 1440 900
capture "02_welcome_pdpa_mobile"  "$BASE_URL/welcome" 390 844

# 3. Rooms Lobby (NFR-US-01, NFR-US-05, NFR-PF-01)
capture "03_rooms_desktop" "$BASE_URL/rooms" 1440 900
capture "03_rooms_mobile"  "$BASE_URL/rooms" 390 844

# 4. Timer Interface (NFR-US-02, NFR-US-03: Work/Rest Feedback)
capture "04_timer_desktop" "$BASE_URL/timer" 1440 900
capture "04_timer_mobile"  "$BASE_URL/timer" 390 844

# 5. Leaderboard (NFR-US-04: Navigation & Gamification)
capture "05_leaderboard_desktop" "$BASE_URL/leaderboard" 1440 900
capture "05_leaderboard_mobile"  "$BASE_URL/leaderboard" 390 844

# 6. Account & Dashboard (NFR-SEC-03: Right to Rectification / Manage Account)
capture "06_account_desktop" "$BASE_URL/account" 1440 900

# 7. Admin Portal (NFR-SEC-02: Role-Based Access Control)
capture "07_admin_desktop" "$BASE_URL/admin" 1440 900

echo ""
echo "======================================================"
echo " All automated screenshots captured successfully!"
echo " Location: $OUTPUT_DIR"
echo "======================================================"

# Non-Functional Requirements (NFRs) Testing and Evidence Collection Guide

**Document Purpose:** Operational procedure guide for executing automated tests, navigating application interfaces, and capturing required evidence artifacts for the Fisher Timer Non-Functional Requirements (NFRs) report.

---

## 1. System Deployment Procedure

Before collecting evidence or executing benchmark scripts, deploy the Fisher Timer system to the local Kubernetes cluster.

### Prerequisites
- Docker engine active (Docker Desktop or OrbStack, minimum 6 GB RAM allocated).
- CLI utilities installed: `kind`, `kubectl`, `helm`, `pnpm`.

### Commands
```bash
# Navigate to the repository root
cd fishertimer

# Verify secrets file is populated
cp k8s/secrets.env.example k8s/secrets.env

# Deploy local Kubernetes cluster and all services
pnpm k8s:up
```

Verify deployment readiness:
```bash
kubectl -n fishertimer get pods
```
All pods must exhibit status `Running` or `Completed` before proceeding. The application entry point will be reachable at `http://localhost`.

---

## 2. Evidence Collection: Interactive Navigation Helper

An interactive helper script is provided to open each target view directly in your default browser:

```bash
bash tests/nfr/open-evidence-pages.sh
```

### Opening Apple Safari for Cross-Browser Testing
To capture the Safari desktop artifact for NFR-OP-01, invoke:
```bash
open -a Safari http://localhost/rooms
```

---

## 3. Step-by-Step Evidence Collection by Requirement

### NFR-OP-01: Multi-Browser and Cross-Platform Support
- **Objective:** Verify visual and functional parity across major desktop browsers and mobile devices.
- **Execution Steps:**
  1. Open `http://localhost/rooms` in Google Chrome (Desktop). Save capture to `docs/nfr-evidence/screenshots/op01_chrome_desktop.png`.
  2. Open `http://localhost/rooms` in Apple Safari (Desktop) using `open -a Safari http://localhost/rooms`. Save capture to `docs/nfr-evidence/screenshots/op01_safari_desktop.png`.
  3. In Chrome Developer Tools, toggle device emulation (`Cmd + Shift + M`), select **iPhone 14 (390 x 844)**, and save capture to `docs/nfr-evidence/screenshots/op01_mobile_viewport.png`.
  4. Run automated Lighthouse audit:
     ```bash
     bash tests/nfr/run-lighthouse.sh
     ```
  5. Copy the score summary from `docs/nfr-evidence/lighthouse-desktop.html` into Section 1 of the report template.

---

### NFR-PF-01: Request Latency under Normal Load
- **Objective:** Validate that common API endpoints respond within 2,000 ms.
- **Execution Steps:**
  1. Execute the k6 benchmark script:
     ```bash
     # Local k6 installation:
     k6 run tests/nfr/k6-common-endpoints.js

     # Alternative via Docker:
     docker run --rm -i --net=host grafana/k6 run - < tests/nfr/k6-common-endpoints.js
     ```
  2. Copy the resulting summary statistics (specifically `http_req_duration p(95)`) into Section 2 of `NFR_REPORT_TEMPLATE.md`.
  3. In Chrome Developer Tools, navigate to the **Network** tab, refresh `http://localhost/rooms`, and capture the timing waterfall. Save image as `docs/nfr-evidence/screenshots/pf01_network_waterfall.png`.

---

### NFR-PF-02: Study Session State Synchronization
- **Objective:** Confirm participant state updates propagate to peers within 2 seconds.
- **Execution Steps:**
  1. Open two browser windows side by side (Window A: User 1 in a room; Window B: User 2 joining the same room).
  2. Join the session in Window B. Observe Window A update within 2 seconds.
  3. Capture the dual-window interface. Save image to `docs/nfr-evidence/screenshots/pf02_dual_client_sync.png`.
  4. Reference code guarantee in `apps/web/features/session/useRoom.ts` (`const ROOM_REFRESH_MS = 2000;`).

---

### NFR-PF-03: Concurrency Scalability (50 Concurrent Users)
- **Objective:** Demonstrate acceptable performance under 50 simultaneous virtual users.
- **Execution Steps:**
  1. Execute the concurrency test script:
     ```bash
     # Local k6:
     k6 run tests/nfr/k6-concurrency-50vu.js

     # Docker execution:
     docker run --rm -i --net=host grafana/k6 run - < tests/nfr/k6-concurrency-50vu.js
     ```
  2. While the test executes, observe Kubernetes pod resource usage in a separate terminal:
     ```bash
     kubectl -n fishertimer top pods
     ```
  3. Copy terminal outputs from both commands into Section 2 of `NFR_REPORT_TEMPLATE.md`.

---

### NFR-SEC-01: Google OAuth Authentication
- **Objective:** Document authentication mechanisms and session security.
- **Execution Steps:**
  1. Navigate to `http://localhost/signin`. Capture the login screen (`sec01_signin_page.png`).
  2. Click "Sign in with Google". On the Google consent screen, capture the browser address bar displaying `client_id` and parameters (`sec01_oauth_consent.png`).
  3. Complete authentication. In Chrome Developer Tools, navigate to **Application > Storage > Cookies > http://localhost**. Capture the `ft_session` entry confirming `HttpOnly` and `SameSite` flags (`sec01_session_cookie.png`).

---

### NFR-SEC-02: Role-Based Access Control (RBAC)
- **Objective:** Validate that administrative endpoints reject unauthenticated or non-admin roles.
- **Execution Steps:**
  1. Run the automated RBAC verification script:
     ```bash
     bash tests/nfr/rbac-security-test.sh
     ```
  2. Copy the assertion output into Section 3 of `NFR_REPORT_TEMPLATE.md`.
  3. Log in with a standard customer account and attempt to navigate to `http://localhost/admin`. Capture the access restriction state (`sec02_customer_blocked.png`).
  4. Log in with an administrator account and navigate to `http://localhost/admin`. Capture the administrative interface (`sec02_admin_dashboard.png`).

---

### NFR-SEC-03 and NFR-CUL-02: PDPA Compliance and Consent
- **Objective:** Validate compliance with Personal Data Protection Act requirements.
- **Execution Steps:**
  1. Navigate to `http://localhost/welcome`.
  2. Click the "Privacy Notice" link to open the policy dialogue.
  3. Capture the modal displaying the personal data collection terms and statutory references (`sec03_pdpa_modal.png`).
  4. Navigate to `http://localhost/account`. Capture the display name modification interface verifying the Right to Rectification under Section 35 (`cul02_profile_rectification.png`).

---

### NFR-CUL-01: English as Default Language
- **Objective:** Confirm UI text and HTML metadata default to English.
- **Execution Steps:**
  1. Navigate to `http://localhost/leaderboard`. Capture the full view verifying English labels across navigation and content (`cul01_english_ui.png`).
  2. Confirm `apps/web/app/layout.tsx` contains `<html lang="en">`.

---

### NFR-US-01: Responsive User Interface
- **Objective:** Demonstrate responsive design adaptation between desktop and mobile.
- **Execution Steps:**
  1. Capture `http://localhost/rooms` at 1440x900 resolution (`us01_desktop_layout.png`).
  2. Capture `http://localhost/rooms` under mobile device emulation at 390x844 resolution (`us01_mobile_layout.png`).

---

### NFR-US-02: Minimal Interaction Timer Operations
- **Objective:** Demonstrate that core timer state changes require single-click or keyboard interactions.
- **Execution Steps:**
  1. Navigate to `http://localhost/timer`.
  2. Capture the timer controls showing direct action buttons (`us02_timer_controls.png`).

---

### NFR-US-03: Visual Feedback for Work and Rest Periods
- **Objective:** Demonstrate unambiguous distinction between active study and rest periods.
- **Execution Steps:**
  1. Start an active work cycle. Capture the interface showing the amber palette, "WORK" badge, and fishing sprite (`us03_work_state.png`).
  2. Transition to a rest period. Capture the interface showing the cyan palette, "REST" badge, and resting sprite (`us03_rest_state.png`).

---

### NFR-US-04: Intuitive Navigation Across Core Features
- **Objective:** Verify global navigation availability across major features.
- **Execution Steps:**
  1. Capture the top-level navigation bar showing direct routes to Rooms, Timer, Leaderboard, and Account (`us04_global_navigation.png`).

---

### NFR-US-05: Minimal Steps to Create or Join a StudySession
- **Objective:** Measure step counts required to join or initialize a room.
- **Execution Steps:**
  1. From `http://localhost/rooms`, highlight the "Join" button on an existing room card (`us05_join_session.png`).
  2. Click "Create Room" to open the creation dialogue. Capture the form demonstrating that only room name input is required (`us05_create_modal.png`).

---

## 4. Final Report Assembly

After collecting all screenshots into `docs/nfr-evidence/screenshots/` and executing the test scripts:
1. Open `docs/nfr-evidence/NFR_REPORT_TEMPLATE.md`.
2. Replace all placeholder text blocks `[INSERT ... HERE]` with the actual test outputs.
3. Review image references to ensure all figures display correctly.
4. Export or commit the completed document for final submission.

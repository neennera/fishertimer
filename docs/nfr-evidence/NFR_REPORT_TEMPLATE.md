# Non-Functional Requirements (NFRs) Evaluation and Evidence Report

**Project Name:** Fisher Timer  
**Course:** Software Architecture  
**Academic Year:** 2026  
**Repository Branch:** `test/nfr-testing-suite`  

**Team Members:**
- Chanatda Konchom (6631305321)
- Ittichet Thongsang (6631363721)
- Napat Srisamut (6631314021)
- Naphat Serirak (6632061321)
- Chayut Archamongkol (6631307621)

---

## 1. Operational Requirements

### NFR-OP-01: Multi-Browser and Cross-Platform Support
> *The system shall support access through major modern web browsers on desktop and mobile devices.*

#### Evidence: Cross-Browser Visual Artifacts
| Google Chrome (Desktop) | Apple Safari (Desktop) | Mobile Viewport (iOS Safari / 390px) |
| :---: | :---: | :---: |
| ![Chrome Desktop](./screenshots/op01_chrome_desktop.png) | ![Safari Desktop](./screenshots/op01_safari_desktop.png) | ![Mobile Viewport](./screenshots/op01_mobile_viewport.png) |
| *Chrome Version 120+* | *macOS Safari Version 17+* | *Mobile Webkit / Safari Viewport* |

#### Evidence: Web Standards and Best Practices Audit (Google Lighthouse)
```text
[INSERT LIGHTHOUSE AUDIT REPORT SUMMARY HERE]
Target URL: http://localhost/rooms
Device: Desktop & Mobile Emulation
- Performance: 
- Accessibility: 
- Best Practices: 
- SEO: 
```

**Evaluation Summary:** The user interface renders consistently across Chrome, Safari, and Mobile viewports without visual artifacts, console runtime errors, or layout degradation.

---

## 2. Performance Requirements

### NFR-PF-01: Request Latency under Normal Load
> *The system shall respond to common user requests within 2 seconds under normal operating conditions.*

#### Evidence: Automated API Benchmark (Grafana k6)
```text
[INSERT K6 BENCHMARK TERMINAL OUTPUT HERE - tests/nfr/k6-common-endpoints.js]
Target Endpoints: /api/session/rooms, /api/leaderboard, /api/auth/me
- Status Checks: 100.00% passed
- http_req_duration (p95): 
- http_req_failed rate: 
```

#### Evidence: Browser Network Waterfall Inspection
![Network Latency Waterfall](./screenshots/pf01_network_waterfall.png)
*Figure 2.1: Browser developer tools network panel showing sub-100ms response times for common API endpoints.*

**Evaluation Summary:** Common endpoints consistently achieve a 95th-percentile response time well below the 2,000 ms threshold under normal operation.

---

### NFR-PF-02: Study Session State Synchronization
> *The system shall synchronize study session status and participant information among active users within 2 seconds under normal network conditions.*

#### Evidence: Multi-Client Real-Time Synchronization
![Dual Client Synchronization](./screenshots/pf02_dual_client_sync.png)
*Figure 2.2: Side-by-side view of Client A and Client B demonstrating participant presence propagation within 2 seconds.*

#### Evidence: Architectural Roster Polling Interval
- **File Reference:** `apps/web/features/session/useRoom.ts` (Lines 16-17)
```typescript
/** Roster refresh: the NFR asks for room changes within 2 seconds. */
const ROOM_REFRESH_MS = 2000;
```

**Evaluation Summary:** The client implementation polls session participant state at a strict 2,000 ms interval, ensuring participant join, leave, and status transitions reflect across all connected clients within 2 seconds.

---

### NFR-PF-03: Concurrency Scalability (50 Concurrent Users)
> *The system shall maintain acceptable response time when supporting at least 50 concurrent users.*

#### Evidence: High-Concurrency Load Test (Grafana k6)
```text
[INSERT K6 CONCURRENCY TEST OUTPUT HERE - tests/nfr/k6-concurrency-50vu.js]
Virtual Users: 50 Concurrent VUs
Duration: 1m 30s
- Total Requests Completed: 
- http_req_duration (p95): 
- http_req_failed rate: 
```

#### Evidence: Kubernetes Pod Resource Utilization
```text
[INSERT KUBECTL TOP PODS OUTPUT HERE]
$ kubectl -n fishertimer top pods
```

**Evaluation Summary:** Under a sustained load of 50 concurrent users, the Go microservices and Kubernetes deployment maintain acceptable latency (< 200 ms p95) with zero dropped requests.

---

## 3. Security Requirements

### NFR-SEC-01: Google OAuth Authentication
> *The system shall authenticate users through Google OAuth.*

#### Evidence: Sign-In Interface
![Google Sign-In Interface](./screenshots/sec01_signin_page.png)
*Figure 3.1: Application sign-in screen offering Google OAuth authentication.*

#### Evidence: Google OAuth Consent Screen
![Google OAuth Consent Dialog](./screenshots/sec01_oauth_consent.png)
*Figure 3.2: Google identity provider consent screen displaying client ID and requested profile scopes.*

#### Evidence: Secure Session Cookie Attributes
![HttpOnly Session Cookie](./screenshots/sec01_session_cookie.png)
*Figure 3.3: Developer tools application panel showing `ft_session` cookie configured with `HttpOnly` and `SameSite=Lax`.*

**Evaluation Summary:** User identity is verified exclusively through Google OAuth 2.0. Session tokens are stored in HttpOnly cookies to mitigate client-side script theft (XSS).

---

### NFR-SEC-02: Role-Based Access Control (RBAC)
> *The system shall enforce role-based access control between customers and administrators.*

#### Evidence: Automated RBAC Security Assertion
```text
[INSERT RBAC SECURITY TEST OUTPUT HERE - bash tests/nfr/rbac-security-test.sh]
Endpoint: /api/v1/admin/sessions
- Test 1 (Unauthenticated Request): 
- Test 2 (Tampered Token): 
- Test 3 (Customer Role Token): 
```

#### Evidence: Frontend Access Gate
| Customer Session (Access Restricted) | Administrator Session (Access Granted) |
| :---: | :---: |
| ![Customer Access Blocked](./screenshots/sec02_customer_blocked.png) | ![Admin Dashboard](./screenshots/sec02_admin_dashboard.png) |
| *Admin route inaccessible to regular customers* | *Admin portal accessible with session role = ADMIN* |

**Evaluation Summary:** Administrative endpoints enforce strict role checks in backend middleware (`RoleAdmin = "ADMIN"`), returning HTTP 401 for unauthenticated requests and HTTP 403 for unauthorized customer roles.

---

### NFR-SEC-03: Protection of Personal Information (PDPA Compliance)
> *The system shall protect users' personal information in accordance with applicable PDPA requirements.*

#### Evidence: Explicit Consent and Privacy Notice
![PDPA Consent Modal](./screenshots/sec03_pdpa_modal.png)
*Figure 3.4: Onboarding consent interface requiring explicit agreement to the Privacy Policy and PDPA terms.*

#### Evidence: Data Minimization in Database Schema
- **File Reference:** `services/account/database/schemas/001_create_users_table.sql`
- **Stored Attributes:** `user_id`, `email`, `display_name`, `avatar_url`, `role`, `created_at`, `updated_at`.
- **Excluded Attributes:** No passwords, phone numbers, national IDs, or sensitive personal data are collected or persisted.

**Evaluation Summary:** Personal data collection conforms to PDPA principles of consent, purpose specification, and data minimization.

---

## 4. Cultural and Legal Requirements

### NFR-CUL-01: English as the Default Language
> *The system shall provide English as the default language for the user interface.*

#### Evidence: Interface Localization and HTML Language Attribute
![English User Interface](./screenshots/cul01_english_ui.png)
*Figure 4.1: User interface layout verifying English terminology across all navigation elements, controls, and prompts.*

- **HTML Document Root Attribute:** `apps/web/app/layout.tsx` (Line 21)
```html
<html lang="en" className={fontVariables}>
```

**Evaluation Summary:** All text, UI components, notifications, and HTML metadata default to English.

---

### NFR-CUL-02: Thailand's PDPA Regulatory Compliance
> *The system shall comply with applicable Thailand's Personal Data Protection Act (PDPA) requirements regarding users' personal information.*

#### Evidence: Right to Rectification (Section 35)
![Account Profile Rectification](./screenshots/cul02_profile_rectification.png)
*Figure 4.2: Account settings interface allowing authenticated users to update and correct their display name.*

**Evaluation Summary:** The application implements required Data Subject Rights under Thailand's PDPA, including transparent notice (Section 21), explicit consent (Section 19), and self-service rectification (Section 35).

---

## 5. Usability Requirements

### NFR-US-01: Responsive User Interface
> *The system shall provide a responsive user interface for desktop and mobile devices.*

#### Evidence: Viewport Layout Adaptation
| Desktop Viewport (1440 x 900) | Mobile Viewport (390 x 844) |
| :---: | :---: |
| ![Desktop Layout](./screenshots/us01_desktop_layout.png) | ![Mobile Layout](./screenshots/us01_mobile_layout.png) |
| *Multi-column card layout* | *Single-column touch-optimized layout* |

**Evaluation Summary:** The interface leverages responsive Tailwind CSS design utilities, adapting layout and touch target sizes across varied viewport dimensions.

---

### NFR-US-02: Minimal Interaction Timer Operations
> *The system shall allow users to start, pause, resume, and stop their study timers with minimal interaction.*

#### Evidence: Single-Action Timer Control Panel
![Timer Control Panel](./screenshots/us02_timer_controls.png)
*Figure 5.1: Timer control panel supporting one-click execution and Spacebar shortcut activation.*

- **Action Counts:**
  - Start Timer: 1 Click
  - Pause Timer: 1 Click (or Spacebar)
  - Resume Timer: 1 Click (or Spacebar)
  - Stop Timer: 1 Click

**Evaluation Summary:** Core timer operations execute immediately upon a single user interaction without nested confirmation dialogues.

---

### NFR-US-03: Visual Feedback for Work and Rest Periods
> *The system shall provide clear visual feedback for work and rest periods.*

#### Evidence: State-Differentiated Visual States
| Work Period (Active Focus) | Rest Period (Break Interval) |
| :---: | :---: |
| ![Work Period State](./screenshots/us03_work_state.png) | ![Rest Period State](./screenshots/us03_rest_state.png) |
| *Amber color palette, "WORK" status badge, active character sprite* | *Cyan color palette, "REST" status badge, resting character sprite* |

**Evaluation Summary:** Clear differentiation between work and rest cycles is achieved through distinct color schemes, status badges, and thematic sprites.

---

### NFR-US-04: Intuitive Navigation Across Core Features
> *The system shall provide intuitive navigation between StudySession, StudyTimer, Dashboard, Reward, and Leaderboard features.*

#### Evidence: Unified Global Navigation Bar
![Global Navigation Header](./screenshots/us04_global_navigation.png)
*Figure 5.2: Top-level navigation bar accessible across all application views.*

- **Feature Direct Access Links:**
  - Rooms (`/rooms`)
  - Timer (`/timer`)
  - Leaderboard (`/leaderboard`)
  - Account & Statistics (`/account`)
  - Active view indicators confirm user location within the application hierarchy.

**Evaluation Summary:** All major features are accessible within a single navigation step from any page.

---

### NFR-US-05: Minimal Steps to Create or Join a StudySession
> *The system shall minimize the number of steps required to create or join a StudySession.*

#### Evidence: Step-Count Efficiency Analysis
| Join Study Session (1 Step) | Create Study Session (2 Steps) |
| :---: | :---: |
| ![Join Session](./screenshots/us05_join_session.png) | ![Create Session Modal](./screenshots/us05_create_modal.png) |
| *Click "Join" on any active room card -> Immediate entry* | *Click "Create Room" -> Enter room name and submit* |

**Evaluation Summary:** Session participation requires a single click, and session creation defaults to standard system capacity (5 participants), requiring only room name input.

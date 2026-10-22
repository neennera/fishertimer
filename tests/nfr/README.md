# Fisher Timer: NFR Testing & Evidence Guide 🎣⏱️

คู่มือการรันสคริปต์ทดสอบและตัวช่วยเปิดหน้าเว็บสำหรับเก็บหลักฐาน Non-Functional Requirements (NFRs) ทั้ง 14 ข้อ

---

## 🧭 1. สคริปต์ช่วยเปิดหน้าเว็บเพื่อกดแคปภาพหน้าจอ (Interactive Navigator)

แทนที่จะต้องมานั่งพิมพ์ URL ทีละหน้า ให้รันสคริปต์นี้เพื่อเปิดหน้าเว็บที่ต้องการลงในเบราว์เซอร์อัตโนมัติ พร้อมคำแนะนำว่าหน้านั้นต้องแคปจุดไหนบ้าง:

```bash
bash tests/nfr/open-evidence-pages.sh
```

**เมนูที่มีให้เลือก:**
1. `NFR-SEC-01`: Sign In & Google OAuth Consent Screen
2. `NFR-SEC-03`: Welcome & PDPA Consent Modal
3. `NFR-US-05`: Rooms Lobby (1-click Join & 2-step Create)
4. `NFR-US-02 / US-03`: Timer Control Panel (Work vs Rest Feedback)
5. `NFR-US-04`: Leaderboard & Global Header
6. `NFR-SEC-03`: Account Page (Edit Display Name)
7. `NFR-SEC-02`: Admin Portal & Moderation Logs
8. `NFR-OP-01 / US-01`: Mobile Responsive Viewport (390px)
9. หรือกด `A` เพื่อเปิดทุกหน้าพร้อมกันในครั้งเดียว!

---

## 🧪 2. คำสั่งรันสคริปต์ทดสอบ (Terminal Commands & Evidence)

### A. ทดสอบ Request Latency < 2 วินาที [NFR-PF-01]
- **ไฟล์สคริปต์:** `tests/nfr/k6-common-endpoints.js`
- **เครื่องมือ:** Grafana k6 (Open-source)
- **คำสั่งรัน:**
  ```bash
  # วิธีที่ 1: ถ้าลง k6 ในเครื่องแล้ว (brew install k6)
  k6 run tests/nfr/k6-common-endpoints.js

  # วิธีที่ 2: รันผ่าน Docker ได้ทันที (ไม่ต้องติดตั้ง k6 ลงเครื่อง)
  docker run --rm -i --net=host grafana/k6 run - < tests/nfr/k6-common-endpoints.js
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - สรุปสถิติ `http_req_duration p(95)` (ต้องได้ < 2000ms โดยระบบจริงจะได้ ~20-50ms)
  - `http_req_failed: 0.00%`

---

### B. ทดสอบ 50 Concurrent Users Scalability [NFR-PF-03]
- **ไฟล์สคริปต์:** `tests/nfr/k6-concurrency-50vu.js`
- **เครื่องมือ:** Grafana k6
- **คำสั่งรัน:**
  ```bash
  # วิธีที่ 1: ใช้ k6 ในเครื่อง
  k6 run tests/nfr/k6-concurrency-50vu.js

  # วิธีที่ 2: ผ่าน Docker
  docker run --rm -i --net=host grafana/k6 run - < tests/nfr/k6-concurrency-50vu.js
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - ข้อความแสดง `50 VUs completed`
  - ตารางสรุป Response Time ภายใต้โหลด 50 คนพร้อมกัน (p95 ยังคง < 2000ms)

---

### C. ทดสอบความปลอดภัย Role-Based Access Control (RBAC) [NFR-SEC-02]
- **ไฟล์สคริปต์:** `tests/nfr/rbac-security-test.sh`
- **เครื่องมือ:** Bash + cURL
- **คำสั่งรัน:**
  ```bash
  bash tests/nfr/rbac-security-test.sh
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - ผลลัพธ์แสดง `Test 1: HTTP 401 Unauthorized PASS`
  - `Test 2: Tampered Token HTTP 401 PASS`
  - `Test 3: Customer Role HTTP 403 Forbidden PASS`

---

### D. ตรวจสอบ Mobile & Web Standards (Google Lighthouse) [NFR-OP-01, US-01]
- **ไฟล์สคริปต์:** `tests/nfr/run-lighthouse.sh`
- **เครื่องมือ:** Google Lighthouse
- **คำสั่งรัน:**
  ```bash
  bash tests/nfr/run-lighthouse.sh
  ```
- **สิ่งที่ต้องแคป:**
  - ไฟล์ HTML รายงานที่เซฟไว้ใน `docs/nfr-evidence/` หรือเปิดดูในเบราว์เซอร์ แสดงคะแนน Performance & Best Practices > 90

---

## 📄 3. นำภาพและผลรันไปวางในเล่มรายงาน
นำภาพ Screenshot และข้อความ Terminal Output ที่แคปได้ ไปแปะลงในไฟล์:
👉 **[`docs/nfr-evidence/NFR_REPORT_TEMPLATE.md`](../docs/nfr-evidence/NFR_REPORT_TEMPLATE.md)**

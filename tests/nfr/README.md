# Fisher Timer: NFR Testing & Evidence Guide 🎣⏱️

คู่มือการเตรียมระบบ การรันสคริปต์ทดสอบ และตัวช่วยเปิดหน้าเว็บสำหรับเก็บหลักฐาน Non-Functional Requirements (NFRs) ทั้ง 14 ข้อ

---

## 🛠️ 0. สิ่งที่ต้องทำก่อนเริ่มรันการทดสอบ (Prerequisites & Setup)

ก่อนจะเริ่มรันสคริปต์ทดสอบ (โดยเฉพาะชุด k6 Performance Tests) **ระบบ Web & Microservices ต้องเปิดใช้งานและตอบสนองที่ `http://localhost` ก่อนเสมอ** เพื่อไม่ให้การทดสอบล้มเหลวจากการหา Service ไม่เจอ:

### วิธีเปิดระบบ (เลือกวิธีใดวิธีหนึ่งตามสภาพแวดล้อมของคุณ):

* **ทางเลือกที่ 1: รันบน Local Kubernetes ด้วย Kind (แนะนำ)**
  ```bash
  # บูต Kind cluster, Traefik Gateway, Build images และ Apply manifests
  pnpm k8s:up

  # ตรวจสอบว่าทั้ง 15 Pods ใน namespace 'fishertimer' มีสถานะ 1/1 Running ครบแล้ว
  kubectl --context kind-fishertimer -n fishertimer get pods
  ```

* **ทางเลือกที่ 2: รันผ่าน Docker Compose**
  ```bash
  # บูตทุก Service และ Databases ทั้งหมด
  pnpm docker:up

  # ตรวจสอบสถานะ Containers
  docker compose ps
  ```

### การตรวจสอบความพร้อมเบื้องต้น (Pre-flight Connectivity Check):
ทดสอบยิง curl เพื่อยืนยันว่า Traefik Gateway และ Services พร้อมรับโหลด:
```bash
curl -I http://localhost/api/session/active
# ต้องได้ HTTP 200 OK ก่อนเริ่มรันชุดทดสอบ
```

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

## 🧪 2. คำสั่งรันสคริปต์ทดสอบ Performance ด้วย `run-k6.sh`

เพื่อความสะดวกและไม่ต้องจำคำสั่งยาวๆ เรามีสคริปต์ **`tests/nfr/run-k6.sh`** ตัวเดียวที่จัดการ:
* ตรวจสอบ Health check ของ `http://localhost` ให้อัตโนมัติก่อนเริ่มเทส
* ตรวจหาเครื่องมือในเครื่องอัตโนมัติ (ถ้าระบบมี `k6` จะรันแบบ Native ทันที แต่ถ้าไม่มีจะรันผ่าน `Docker` ให้อัตโนมัติโดยไม่ต้องติดตั้งเพิ่ม)

```
tests/nfr/
├── run-k6.sh                        <-- สคริปต์ตัวกลางสำหรับสั่งรัน
└── k6/                              <-- โฟลเดอร์เก็บโค้ดสคริปต์ k6 ทั้งหมด
    ├── k6-common-endpoints.js
    ├── k6-concurrency-50vu.js
    ├── k6-stress-test.js
    └── k6-spike-test.js
```

### สั่งรันผ่าน Interactive Menu:
```bash
bash tests/nfr/run-k6.sh
# หรือใช้ shorthand
pnpm test:k6
```

---

### หรือสั่งรันรายตัวตามต้องการ:

#### A. ทดสอบ Request Latency < 2 วินาที (Baseline SLA) [NFR-PF-01]
- **ไฟล์สคริปต์:** `tests/nfr/k6/k6-common-endpoints.js`
- **คำสั่งรันสั้นๆ:**
  ```bash
  bash tests/nfr/run-k6.sh baseline
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - สถิติ `http_req_duration p(95)` (< 2000ms)
  - `http_req_failed: 0.00%`

---

#### B. ทดสอบ 50 Concurrent Users Scalability (Load Test) [NFR-PF-03]
- **ไฟล์สคริปต์:** `tests/nfr/k6/k6-concurrency-50vu.js`
- **คำสั่งรันสั้นๆ:**
  ```bash
  bash tests/nfr/run-k6.sh concurrency
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - ข้อความแสดง `50 VUs completed`
  - ตารางสรุป Response Time ภายใต้โหลด 50 คนพร้อมกัน (p95 ยังคง < 2000ms)

---

#### C. ทดสอบความทนทานและการลดทอนประสิทธิภาพ (Stress Test: 150 VUs)
- **ไฟล์สคริปต์:** `tests/nfr/k6/k6-stress-test.js`
- **วัตถุประสงค์:** ผลักดันโหลดขึ้นไป 3 เท่าของโหลดปกติ (0 → 50 → 100 → 150 VUs รวม 90 วินาที) ตรวจสอบความทนทานและ Recovery
- **คำสั่งรันสั้นๆ:**
  ```bash
  bash tests/nfr/run-k6.sh stress
  # หรือใช้ shorthand
  pnpm test:stress
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - ตารางสรุป HTTP Request Duration และ Error Rate ที่โหลดสูงสุด 150 VUs (Threshold: error rate < 5%)

---

#### D. ทดสอบการรับมือคลื่น Traffic เฉียบพลัน (Spike Test: 10 → 150 VUs)
- **ไฟล์สคริปต์:** `tests/nfr/k6/k6-spike-test.js`
- **วัตถุประสงค์:** จำลองสถานการณ์ Traffic กระชากฉับพลันจาก 10 เป็น 150 VUs ภายใน 5 วินาที เพื่อดู Connection Queue และการดีดกลับสู่ Baseline Latency
- **คำสั่งรันสั้นๆ:**
  ```bash
  bash tests/nfr/run-k6.sh spike
  # หรือใช้ shorthand
  pnpm test:spike
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - ผลลัพธ์แสดงการคงอยู่ของระบบโดยไม่มี Connection Reset เกินเกณฑ์ (Threshold: error rate < 5%)

---

#### E. รัน Performance Tests ทั้งหมดต่อเนื่องกันในครั้งเดียว:
```bash
bash tests/nfr/run-k6.sh all
```

---

## 🔒 3. ทดสอบความปลอดภัย Role-Based Access Control (RBAC) [NFR-SEC-02]
- **ไฟล์สคริปต์:** `tests/nfr/rbac-security-test.sh`
- **คำสั่งรัน:**
  ```bash
  bash tests/nfr/rbac-security-test.sh
  ```
- **สิ่งที่ต้องแคปใน Terminal:**
  - ผลลัพธ์แสดง `Test 1: HTTP 401 Unauthorized PASS`
  - `Test 2: Tampered Token HTTP 401 PASS`
  - `Test 3: Customer Role HTTP 403 Forbidden PASS`
  - `Test 4: Admin Role HTTP 200 OK PASS`

---

## 📱 4. ตรวจสอบ Mobile & Web Standards (Google Lighthouse) [NFR-OP-01, US-01]
- **ไฟล์สคริปต์:** `tests/nfr/run-lighthouse.sh`
- **คำสั่งรัน:**
  ```bash
  bash tests/nfr/run-lighthouse.sh
  ```
- **สิ่งที่ต้องแคป:**
  - ไฟล์ HTML รายงานที่เซฟไว้ใน `docs/nfr-evidence/` หรือเปิดดูในเบราว์เซอร์ แสดงคะแนน Performance & Best Practices > 90

---

## 📄 5. นำภาพและผลรันไปวางในเล่มรายงาน
นำภาพ Screenshot และข้อความ Terminal Output ที่แคปได้ ไปแปะลงในไฟล์:
👉 **[`docs/nfr-evidence/NFR_REPORT_TEMPLATE.md`](../docs/nfr-evidence/NFR_REPORT_TEMPLATE.md)**

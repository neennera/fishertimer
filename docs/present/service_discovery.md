# Service Discovery (Consul) — การทำงานและการ Setup

เอกสารนี้อธิบายว่า service discovery ใน Fisher Timer ทำงานยังไง และต้อง setup อะไรบ้าง
รายละเอียดเชิงสถาปัตยกรรมอยู่ที่ [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md) ส่วนที่ 7
และรายละเอียดของ library อยู่ที่ [`pkg/discovery/README.md`](../../pkg/discovery/README.md)

---

## 1. ทำไมต้องมี

ปกติแต่ละ service รู้ address ของ service อื่นจาก env เช่น `REWARD_SERVICE_URL=http://reward:8085`
ซึ่งใช้ได้ แต่:

| ปัญหา | แบบ static (เดิม) | แบบ Consul |
| :--- | :--- | :--- |
| รู้ไหมว่า instance ไหน **healthy** | ไม่รู้ (DNS ตอบแค่ว่า container อยู่ที่ไหน) | รู้ — Consul เช็ก health ทุก 5 วินาที |
| เห็นภาพรวมว่าอะไรกำลังรันอยู่ | ต้องดู `docker ps` | ดูได้ที่ http://localhost:8500 |
| มีหลาย instance ของ service เดียวกัน | พึ่ง DNS | round-robin ระหว่าง instance ที่ healthy |
| service ตายกลางทาง | request ยังถูกส่งไปที่เดิม | หลุดจากรายการภายในไม่กี่วินาที |

Service discovery เป็น **ตัวเลือก** ค่า default ยังเป็นแบบ static เหมือนเดิมทุกอย่าง

---

## 2. ภาพรวมการทำงาน

```
 service เริ่มทำงาน --- register (ชื่อ, ip:port, health check) ---> Consul
 service ปิดตัว    --- deregister ------------------------------> Consul
 Consul ----- เช็ก /health (HTTP) หรือ TCP ทุก 5 วินาที ------> ทุก instance

 caller ---- "reward ตัวไหน healthy?" (cache 3 วินาที) --------> Consul
 caller ---- ส่ง request ไปยัง 1 ตัว (round-robin) -------------> reward
```

มี 2 ฝั่งที่ทำงานเหมือนกันทุก service (เขียนไว้ครั้งเดียวใน `pkg/discovery`):

### 2.1 ฝั่ง register (Announce)
- ตอน start แต่ละ service ลงทะเบียนตัวเองกับ Consul โดย **ไม่บล็อกการ start**
- พอร์ต HTTP ลงทะเบียนด้วยชื่อ service เช่น `reward` พร้อม health check `GET /health`
- พอร์ต gRPC (study-session, study-timer) ลงทะเบียนเป็น `<ชื่อ>-grpc` พร้อม TCP check
- ลงทะเบียนซ้ำทุก 15 วินาที ดังนั้นถ้า Consul รีสตาร์ท รายการจะกลับมาเอง
- ตอนได้รับ SIGTERM (`docker stop`) จะ deregister ก่อนปิด จึงหายจากรายการทันที
- ถ้าถูก kill แบบไม่ทันปิดตัว health check จะ fail และ Consul เลิกส่ง instance นั้นให้ภายในไม่กี่วินาที

### 2.2 ฝั่ง lookup (Locator)
- **HTTP**: `Locator.Transport` เปลี่ยน host ของแต่ละ request ไปเป็น instance ที่ healthy
  (path และ query ไม่เปลี่ยน) ใช้ทั้งใน reverse proxy ของ api-gateway และ HTTP client ของ
  account / admin / leaderboard / study-timer
- **gRPC**: ใช้ resolver `discovery:///<service>-grpc` พร้อม `round_robin` ใช้ใน
  api-gateway, study-session, study-timer
- ผลที่ได้ cache 3 วินาที ถ้า Consul ตอบไม่ได้จะใช้ผลล่าสุดที่เคยได้

### 2.3 ชื่อ service ที่ลงทะเบียน

| ชื่อ | พอร์ต | ประเภท |
| :--- | :--- | :--- |
| `api-gateway` | 8000 | HTTP |
| `account` | 8082 | HTTP |
| `study-session` / `study-session-grpc` | 8083 / gRPC | HTTP / gRPC |
| `study-timer` / `study-timer-grpc` | 8084 / gRPC | HTTP / gRPC |
| `reward` | 8085 | HTTP |
| `leaderboard` | 8086 | HTTP |
| `admin` | 8087 | HTTP |

---

## 3. Setup

### 3.1 สิ่งที่ต้องมี
- Docker Desktop (หรือ Docker Engine + Compose v2)
- pnpm (ใช้เรียก script) — ไม่ต้องติดตั้ง Go เพื่อรันผ่าน Docker
- Google OAuth client ถ้าต้องการ login (ดูข้อ 3.3)

### 3.2 รัน

```bash
pnpm docker:discovery      # stack ทั้งหมด + Consul
```

ครั้งแรก build นานประมาณ 3-4 นาที ครั้งต่อไปเร็วขึ้นมาก เมื่อเสร็จแล้ว:

| ที่อยู่ | คืออะไร |
| :--- | :--- |
| http://localhost:3000 | เว็บ |
| http://localhost:8500 | Consul UI (ดูหน้า **Services**) |

คำสั่งอื่นที่เกี่ยวข้อง:

```bash
pnpm docker:up             # stack ทั้งหมด ไม่มี Consul (แบบ static)
pnpm docker:dev            # เหมือน docker:up + เปิดพอร์ต DB / gateway / admin ออก localhost
pnpm docker:down           # หยุด
pnpm docker:reset          # หยุดและลบ volume (ข้อมูล DB หาย)
pnpm docker:logs           # ดู log
```

ภายใน `pnpm docker:discovery` คือ `docker-compose.yml` + `docker-compose.discovery.yml`
ไฟล์ที่สองเพิ่ม container `consul` และใส่ env ให้ทั้ง 7 service

### 3.3 Google login
Docker ไม่ใส่ credentials ให้ ต้องตั้งเอง (ครั้งเดียว):

1. `copy .env.docker.example .env.docker` แล้วแก้ `GOOGLE_CLIENT_ID` และ `GOOGLE_CLIENT_SECRET`
2. ใน Google Cloud Console เพิ่ม Authorized redirect URI:
   `http://localhost:3000/api/auth/google/callback`
3. รัน `pnpm docker:discovery` ซ้ำเพื่อให้ container อ่านค่าใหม่ (ไม่ต้อง build ใหม่)

`.env.docker` ถูก git-ignore ห้าม commit

### 3.4 Environment variables

| ตัวแปร | ค่า | ความหมาย |
| :--- | :--- | :--- |
| `DISCOVERY_MODE` | `static` (default) / `consul` | `static` = ไม่ register ไม่ lookup ใช้ address จาก env เหมือนเดิม |
| `CONSUL_HTTP_ADDR` | เช่น `http://consul:8500` | address ของ Consul (ตั้งให้อัตโนมัติใน `docker-compose.discovery.yml`) |
| `SERVICE_ADVERTISE_ADDR` | optional | host ที่ใช้ประกาศใน Consul ถ้าไม่ตั้งจะตรวจจับ IP ของ container เอง |

Address ของ peer เดิม (เช่น `REWARD_SERVICE_URL`, `TIMER_GRPC_TARGET`) **ยังต้องตั้งไว้เหมือนเดิม**
เพราะเป็นค่า fallback ตอนหา instance จาก Consul ไม่ได้

---

## 4. ตรวจสอบว่าทำงาน

1. เปิด http://localhost:8500 → **Services** ต้องเห็น 9 รายการเป็นสีเขียว
   (`account`, `admin`, `api-gateway`, `leaderboard`, `reward`, `study-session`,
   `study-session-grpc`, `study-timer`, `study-timer-grpc`)
2. ทดสอบหยุด service:
   ```bash
   docker stop fishertimer-reward-1    # reward หายจาก Consul ทันที
   docker start fishertimer-reward-1   # กลับมาเมื่อ healthy
   ```
3. พิสูจน์ว่า lookup ทำงานจริง (ไม่ได้ใช้แค่ address ใน env): ใส่ address ผิดโดยตั้งใจใน
   `.env.docker` เช่น `REWARD_SERVICE_URL=http://nope:8085` แล้วรัน `pnpm docker:discovery`
   ระบบยังต้องทำงานได้ แต่ถ้ารันแบบ `pnpm docker:up` (ไม่มี Consul) จะได้ 502/503
   — **ลบบรรทัดนั้นออกหลังทดสอบ**

---

## 5. เมื่อเกิดปัญหา

| สถานการณ์ | พฤติกรรม |
| :--- | :--- |
| Consul ล่ม แต่ peer เคยถูกเรียกมาแล้ว | ใช้ instance ล่าสุดที่รู้จัก ระบบยังทำงานต่อ |
| Consul ล่ม และ peer ยังไม่เคยถูกเรียกมาก่อน | ใช้ address ที่ตั้งไว้ใน env (fallback) |
| Consul กลับมา | service register ซ้ำเองภายใน 15 วินาที |
| service ไม่ได้ register (เช่นรันบนเครื่องใน hybrid mode) | ใช้ address ที่ตั้งไว้ใน env |
| container ถูก kill | หลุดจากรายการภายในไม่กี่วินาที (Consul ลบทิ้งทั้งหมดหลังเกิน 1 นาที) |
| service start ก่อน Consul พร้อม | start ได้ตามปกติ และ register ทีหลัง (compose รอ Consul healthy อยู่แล้ว) |

ดู log ของการตัดสินใจได้ เช่น `docker logs fishertimer-api-gateway-1 | findstr registry`
จะเห็นข้อความทำนอง `registry unavailable ..., using configured address`

---

## 6. ข้อจำกัด

- Consul เป็น **dev agent ตัวเดียว**: ไม่มี persistence, HA, TLS หรือ ACL เหมาะกับ local เท่านั้น
  (restart แล้วรายการว่าง จนกว่า service จะ register ซ้ำ)
- cache 3 วินาที: request อาจยังถูกส่งไป instance ที่เพิ่งตายได้ชั่วครู่
- เว็บ (Next.js) ไม่ใช้ registry เพราะ `rewrites()` ถูกฝังตอน build
- Kubernetes ใช้ DNS ของ cluster อยู่แล้ว ไม่ต้องเปิดโหมดนี้ (ค่า default `static`)
- Docker Hub rate limit (429) ตอน pull image ครั้งแรกเป็นปัญหาที่เจอได้ ลองใหม่ภายหลังหรือ `docker login`

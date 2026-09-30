-- ==============================================================================
-- [REMARK: TEMPORARY MOCK SEED ONLY]
-- Study Timer Service: Phase 2 W1 Local Testing Fixture
-- Database: timer_db
-- ==============================================================================
-- REMARK / CONTEXT:
-- • สถานะ: TEMPORARY MOCK FIXTURE (ใช้เฉพาะการทดสอบ Local ของ Role B)
-- • รอเชื่อมต่อกับ: Study Session Service (Role A) ใน Phase 2 Week 2 (W2)
-- • คำอธิบาย:
--   ในระบบจริง (Production / Integrated System) ตาราง timers และ cycles จะถูกสร้าง
--   แบบ Dynamic เมื่อผู้ใช้สร้าง/เข้าร่วมห้องและกดเริ่มนาฬิกา (gRPC StartTimer)
--   ไม่ใช่การ Seed ไว้ล่วงหน้าแบบคงที่
--
-- • วัตถุประสงค์ของ Mock Seed นี้:
--   จำลองห้องเรียนและ Timer ที่เปิดอยู่ (OPEN) พร้อม Work Cycle ที่กำลังรัน (RUNNING)
--   เพื่อให้สามารถทดสอบการทำงานของ RabbitMQ Event Consumer (เช่น session.participant.left
--   และ session.ended) รวมถึงคำนวณเวลาและตัดจบรอบ (Finalize) ได้ทันทีก่อนที่
--   Study Session Service จะพัฒนาระบบ JoinSession เสร็จสมบูรณ์
--
-- • ข้อควรระวัง:
--   สามารถ Wipe หรือล้างข้อมูลชุดนี้ทิ้งได้เมื่อเชื่อมต่อระหว่าง Service ครบถ้วน
-- ==============================================================================

-- 1. Seed user timer in room
INSERT INTO timers (timer_id, session_id, user_id, status, opened_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'OPEN',
    CURRENT_TIMESTAMP
)
ON CONFLICT (session_id, user_id) DO UPDATE
SET status = 'OPEN', finalized_at = NULL;

-- 2. Seed active WORK cycle currently RUNNING
INSERT INTO cycles (cycle_id, timer_id, type, status, duration_sec, started_at, reward_status)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'WORK',
    'RUNNING',
    1500,
    CURRENT_TIMESTAMP,
    'NONE'
)
ON CONFLICT (cycle_id) DO UPDATE
SET status = 'RUNNING', ended_at = NULL;

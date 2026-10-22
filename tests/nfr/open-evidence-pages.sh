#!/usr/bin/env bash
# ==============================================================================
# Fisher Timer: NFR Browser Navigation Helper 🧭
# สคริปต์ช่วยเปิดหน้าเว็บ Fisher Timer บนเบราว์เซอร์ เพื่อให้คุณกดแคปภาพหน้าจอ
# ประกอบ NFR แต่ละข้อได้อย่างสะดวก ไม่ต้องคอยพิมพ์ URL เอง
# ==============================================================================
set -euo pipefail

BASE_URL="${1:-http://localhost}"

# Colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
NC='\033[0m'

clear_screen() {
  printf "\033c"
}

show_menu() {
  clear_screen
  echo -e "${CYAN}====================================================================${NC}"
  echo -e "${CYAN}    🎣 Fisher Timer: NFR Evidence Browser Navigator                 ${NC}"
  echo -e "${CYAN}====================================================================${NC}"
  echo -e "Base URL: ${GREEN}$BASE_URL${NC} (รันระบบด้วย pnpm k8s:up ก่อน)"
  echo ""
  echo -e "เลือกข้อ NFR ที่ต้องการเปิดหน้าเว็บเพื่อแคปภาพหน้าจอ:"
  echo ""
  echo -e "  ${YELLOW}[1]${NC} NFR-SEC-01 : หน้าลงชื่อเข้าใช้ Google OAuth        (${BASE_URL}/signin)"
  echo -e "  ${YELLOW}[2]${NC} NFR-SEC-03 : หน้า Onboarding & PDPA Consent Modal  (${BASE_URL}/welcome)"
  echo -e "  ${YELLOW}[3]${NC} NFR-US-05  : หน้ารวมห้อง Study Rooms (Join/Create) (${BASE_URL}/rooms)"
  echo -e "  ${YELLOW}[4]${NC} NFR-US-02  : หน้าแผงควบคุม Timer (Start/Pause/Stop)(${BASE_URL}/timer)"
  echo -e "  ${YELLOW}[5]${NC} NFR-US-04  : หน้า Leaderboard & Header Navigation (${BASE_URL}/leaderboard)"
  echo -e "  ${YELLOW}[6]${NC} NFR-SEC-03 : หน้า Account (แก้ไข Display Name)     (${BASE_URL}/account)"
  echo -e "  ${YELLOW}[7]${NC} NFR-SEC-02 : หน้า Admin Portal (RBAC Moderation)   (${BASE_URL}/admin)"
  echo -e "  ${YELLOW}[8]${NC} NFR-OP-01  : เปิดหน้า Rooms เพื่อทดสอบ Mobile View (390px)"
  echo -e "  ${YELLOW}[S]${NC} NFR-OP-01  : เปิดหน้า Rooms ใน Apple Safari Desktop (${BASE_URL}/rooms)"
  echo -e "  ${YELLOW}[A]${NC} ${GREEN}เปิดทุกหน้าพร้อมกันใน Tabs ทั้งหมด${NC}"
  echo -e "  ${YELLOW}[Q]${NC} ออกจากโปรแกรม"
  echo ""
  echo -e "${CYAN}--------------------------------------------------------------------${NC}"
}

while true; do
  show_menu
  read -p "พิมพ์ตัวเลือกที่ต้องการ (1-8, A, Q) แล้วกด Enter: " choice
  case "$choice" in
    1)
      echo -e "${GREEN}กำลังเปิดหน้า Sign In...${NC}"
      open "$BASE_URL/signin"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ:${NC}"
      echo "  1. ปุ่ม 'Sign in with Google'"
      echo "  2. เมื่อกดแล้ว แคปหน้าต่าง Google OAuth Consent Screen ที่แสดง Client ID"
      echo "  3. หลังล็อกอิน กด F12 -> Application -> Cookies แคปคุกกี้ 'ft_session' (HttpOnly)"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    2)
      echo -e "${GREEN}กำลังเปิดหน้า Welcome / PDPA Consent...${NC}"
      open "$BASE_URL/welcome"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ:${NC}"
      echo "  1. กล่องติ๊กยอมรับ 'I have read and acknowledge the Privacy Notice...'"
      echo "  2. คลิกที่คำว่า 'Privacy Notice' เพื่อเปิด Modal แล้วแคปหน้าต่างข้อตกลง PDPA เต็มๆ"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    3)
      echo -e "${GREEN}กำลังเปิดหน้ารวมห้อง (Rooms)...${NC}"
      open "$BASE_URL/rooms"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ (NFR-US-05 & NFR-US-01):${NC}"
      echo "  1. ปุ่ม 'Join' บน Room Card (แสดงว่าเข้าห้องได้ใน 1 คลิก)"
      echo "  2. คลิกปุ่ม 'Create Room' แล้วแคปหน้า Modal สร้างห้อง (แสดงว่ากรอกแค่ชื่อห้องแล้วกด Create ได้ทันที)"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    4)
      echo -e "${GREEN}กำลังเปิดหน้า Timer...${NC}"
      open "$BASE_URL/timer"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ (NFR-US-02 & NFR-US-03):${NC}"
      echo "  1. ปุ่มควบคุมเวลา Start, Pause, Resume, Stop (1 คลิกสั่งการได้ทันที)"
      echo "  2. โหมด WORK: ป้ายสีทอง/ส้ม ระบุ 'WORK' พร้อมอนิเมะตกปลา"
      echo "  3. โหมด REST: ป้ายสีฟ้าคราม ระบุ 'REST' พร้อมอนิเมะท่านั่งพัก"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    5)
      echo -e "${GREEN}กำลังเปิดหน้า Leaderboard...${NC}"
      open "$BASE_URL/leaderboard"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ (NFR-US-04 & NFR-PF-01):${NC}"
      echo "  1. แถบ Header ด้านบน (Rooms, Timer, Leaderboard, Account) แสดงการนำทางที่ชัดเจนใน 1 คลิก"
      echo "  2. ตารางอันดับพร้อม Filter: All-Time, Monthly, Weekly"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    6)
      echo -e "${GREEN}กำลังเปิดหน้า Account...${NC}"
      open "$BASE_URL/account"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ (NFR-SEC-03: Right to Rectification):${NC}"
      echo "  1. หน้าโปรไฟล์แสดง Email และ Display Name"
      echo "  2. ช่องแก้ไขชื่อ Display Name พร้อมปุ่ม Save (พิสูจน์การแก้ไขข้อมูลส่วนบุคคลตามสิทธิ PDPA)"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    7)
      echo -e "${GREEN}กำลังเปิดหน้า Admin Portal...${NC}"
      open "$BASE_URL/admin"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ (NFR-SEC-02: RBAC):${NC}"
      echo "  1. หากล็อกอินเป็น User ทั่วไป (CUSTOMER): แคปหน้าที่ถูกบล็อกหรือเด้งออก"
      echo "  2. หากล็อกอินเป็น ADMIN: แคปหน้า Admin Dashboard รายการห้องและแท็บ Audit Logs (/admin/logs)"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    8)
      echo -e "${GREEN}กำลังเปิดหน้า Rooms (สำหรับทดสอบ Responsive Mobile)...${NC}"
      open "$BASE_URL/rooms"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องทำและแคป (NFR-OP-01 & NFR-US-01):${NC}"
      echo "  1. กด F12 หรือ Cmd + Option + I เปิด DevTools"
      echo "  2. กดปุ่ม Device Toggle (Cmd + Shift + M) แล้วเลือก 'iPhone 14' (390 x 844)"
      echo "  3. แคปหน้าจอ Mobile Layout ที่ Room Cards เรียงเป็น Single Column สวยงาม"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    S|s)
      echo -e "${GREEN}กำลังเปิดหน้า Rooms ใน Apple Safari...${NC}"
      open -a Safari "$BASE_URL/rooms" || open "$BASE_URL/rooms"
      echo ""
      echo -e "${MAGENTA}📸 สิ่งที่ต้องแคปหน้าจอ (NFR-OP-01: Safari Support):${NC}"
      echo "  1. แคปหน้าต่าง Safari เต็มจอ"
      echo "  2. บันทึกรูปภาพลงที่: docs/nfr-evidence/screenshots/op01_safari_desktop.png"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    A|a)
      echo -e "${GREEN}กำลังเปิดทุกหน้าพร้อมกันใน Browser Tabs...${NC}"
      open "$BASE_URL/signin"
      open "$BASE_URL/welcome"
      open "$BASE_URL/rooms"
      open "$BASE_URL/timer"
      open "$BASE_URL/leaderboard"
      open "$BASE_URL/account"
      open "$BASE_URL/admin"
      echo "เปิดครบทุกหน้าเรียบร้อยแล้ว!"
      read -p "กด Enter เพื่อกลับเมนูหลัก..." _
      ;;
    Q|q)
      echo "ออกจากโปรแกรมเรียบร้อย"
      exit 0
      ;;
    *)
      echo "ตัวเลือกไม่ถูกต้อง กรุณาลองใหม่"
      sleep 1
      ;;
  esac
done

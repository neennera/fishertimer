# Feature: Session

Study rooms (UC-01 Create, UC-02 Join, UC-03 Leave). Pages: `app/rooms/page.tsx` (the lake: room list + create) and `app/rooms/[id]/page.tsx` (one room: the dock, your timer, Leave). `app/rooms/layout.tsx` keeps the lake scene and header mounted while moving between them.

| File | Role |
| --- | --- |
| `session.api.ts` | The only file that calls the backend for rooms: `/api/session/{active,me,room,create,join,leave,heartbeat}` (gateway → Study Session gRPC) and `/api/timer/room`. Errors carry a stable `code` (`ROOM_FULL`, `ROOM_ENDED`, `ALREADY_IN_ROOM`, ...). Always sends `user_id` / `display_name`; the gateway ignores them when a session cookie is present (they only matter with mock auth). |
| `useActor.ts` | The signed-in user from `lib/auth.ts`. |
| `useRoomList.ts` | Active rooms + the user's current room, refreshed every 4s and on tab focus. |
| `useRoomActions.ts` | Join / create, including the "already in a room" detour (UC-01 E-2, UC-02 E-3): leave the old room, then retry. |
| `useRoom.ts` | One room kept live: roster every 2s (NFR: 2s sync), room timers every 3s, heartbeat every 10s. Reports `removed` with the reason when the user is kicked, timed out, idle or the room closes. |
| `summary.ts` | The leave summary (UC-03 step 5): cycles and focus time from the user's own timer read just before leaving; fish from Reward with `awarded_at` inside the stay. |
| `components/RoomDock.tsx` | The room drawn as a pier: one angler per participant, line and bobber following their timer (`.pixel-dock` / `.pixel-angler` in `pixel.css`), plus the accessible roster list. |
| `components/angler.ts` | Timer state → dock state (`focus`, `paused`, `caught`, `rest`, `idle`) and a stable colour per user. |
| `components/RoomCard.tsx` | One room on the list with seat pips. |
| `components/CreateRoomModal.tsx` | Name + seat picker (1-5, default 5), validated like the server. |
| `components/LeaveRoomDialog.tsx` | Confirm leave; warns when a work cycle would be forfeited (UC-03 S-1). |
| `components/SessionSummaryModal.tsx` | The catch after leaving. |
| `components/AlreadyInRoomModal.tsx` | Leave the current room or cancel. |

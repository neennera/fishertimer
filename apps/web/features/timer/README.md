# Feature: Timer

Contains components, custom React hooks, and view models specific to the timer domain.

| File | Role |
| --- | --- |
| `timer.api.ts` | The only file that calls the backend: `GET /api/timer/state`, `POST /api/timer/{start,pause,resume,reset}`. The gateway turns these into gRPC calls to Study Timer. |
| `useStudyTimer.ts` | Holds one participant's timer. The server owns the time; the hook counts down locally for display and re-reads the server when the tab regains focus or the countdown reaches zero. |
| `TimerPanel.tsx` | Countdown and controls. Shows only the buttons valid for the current status. |

The page is `app/timer/page.tsx`. It currently uses a fixed demo owner
(`session_id` / `user_id` = `demo`); replace it with the real room and
signed-in user once rooms (UC-01/02) exist.

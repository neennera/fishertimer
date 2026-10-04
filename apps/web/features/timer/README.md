# Feature: Timer

Contains components, custom React hooks, and view models specific to the timer domain.

| File | Role |
| --- | --- |
| `timer.api.ts` | The only file that calls the backend: `GET /api/timer/state`, `POST /api/timer/{start,pause,resume,stop,reset,complete,skip-rest,settings}`. The gateway turns these into gRPC calls to Study Timer, acting as the signed-in user. |
| `timer-model.ts` | Pure client model: the local reading of the timer, remaining time, the view (`ready`, `focus`, `focus-paused`, `rest-ready`, `rest`, `rest-paused`, `finishing`, `closed`, from the server's `state`), and `predict()`, which mirrors the server's rules so a press shows before the server answers. |
| `useStudyTimer.ts` | Holds one participant's timer. Presses show instantly (optimistic), are sent one at a time, only the last answer is applied. At zero it asks the server to complete the cycle; it reports a newly completed work block (`completed`) for the catch reveal. Re-reads on load, tab focus and every 15s. |
| `TimerPanel.tsx` | Phase track (Focus / Break), countdown, progress, a length picker before each block and break (UC-05 steps 2 and 10), and per-state controls: Start / Pause / Resume, Reset, Stop, Skip break. Space, R and S shortcuts; the countdown shows in the tab title. |
| `DurationPicker.tsx` | Presets plus - / +, kept inside the server's limits (E-1). |
| `TimerSettingsForm.tsx` | Default lengths (UpdateTimerSetting), in the panel's gear modal and on `/settings`. |
| `CatchReveal.tsx` | After a work block: reels in the fish Reward awarded for it (UC-09 step 9). |

`TimerPanel` runs inside every study room (`app/rooms/[id]/page.tsx`) with the
real room id and signed-in user. `app/timer/page.tsx` is still a standalone
demo with a fixed owner (UUIDs `...0001` / `...0002`).

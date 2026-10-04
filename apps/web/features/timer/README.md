# Feature: Timer

Contains components, custom React hooks, and view models specific to the timer domain.

| File | Role |
| --- | --- |
| `timer.api.ts` | The only file that calls the backend: `GET /api/timer/state`, `POST /api/timer/{start,pause,resume,reset}`. The gateway turns these into gRPC calls to Study Timer. |
| `timer-model.ts` | Pure client model: the local reading of the timer, remaining time, the `ready / running / paused / done` view, and `predict()`, which mirrors the server's rules so a press can be shown before the server answers. |
| `useStudyTimer.ts` | Holds one participant's timer. Presses show instantly (optimistic), are sent one at a time, and only the last answer is applied. The display advances every animation frame; the server is re-read on load, on tab focus and once at zero. |
| `TimerPanel.tsx` | Countdown, progress bar and controls. The primary button keeps its place and width across Start / Pause / Resume; Reset is always present. Space and R shortcuts; the countdown shows in the tab title. |

`TimerPanel` runs inside every study room (`app/rooms/[id]/page.tsx`) with the
real room id and signed-in user. `app/timer/page.tsx` is still a standalone
demo with a fixed owner (UUIDs `...0001` / `...0002`).

import { Header } from "../../components/Header";
import { TimerPanel } from "../../features/timer/TimerPanel";
import type { TimerOwner } from "../../features/timer/timer.api";

// Demo only: the timer is not linked to a signed-in user or a real room yet.
// Once rooms (UC-01/02) land, the room id and the signed-in user id replace
// these fixed values. They are UUIDs because timer_db keys timers by UUID.
const DEMO_OWNER: TimerOwner = {
  sessionId: "00000000-0000-0000-0000-000000000001",
  userId: "00000000-0000-0000-0000-000000000002",
};

export default function TimerPage() {
  return (
    <div className="flex min-h-screen flex-col">
      <div className="pixel-scene" aria-hidden="true" />
      <Header />
      {/* Sits in the sky, above the horizon, so the scene frames the panel. */}
      <main className="flex flex-1 items-start justify-center px-4 pt-[12vh] pb-12">
        <div className="w-full max-w-md">
          <TimerPanel owner={DEMO_OWNER} />
        </div>
      </main>
    </div>
  );
}

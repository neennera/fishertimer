import { Header } from "../../components/Header";
import { TimerPanel } from "../../features/timer/TimerPanel";
import type { TimerOwner } from "../../features/timer/timer.api";

// Demo only: the timer is not linked to a signed-in user or a real room yet.
// Once rooms (UC-01/02) land, the room id and the signed-in user id replace
// these fixed values.
const DEMO_OWNER: TimerOwner = { sessionId: "demo", userId: "demo" };

export default function TimerPage() {
  return (
    <>
      <Header />
      <main className="mx-auto w-full max-w-md px-4 py-12">
        <TimerPanel owner={DEMO_OWNER} />
      </main>
    </>
  );
}

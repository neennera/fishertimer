"use client";

import Link from "next/link";
import { use, useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { PixelAlert } from "../../../components/ui/PixelAlert";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelPanel } from "../../../components/ui/PixelPanel";
import { AlreadyInRoomModal } from "../../../features/session/components/AlreadyInRoomModal";
import { LeaveRoomDialog } from "../../../features/session/components/LeaveRoomDialog";
import { RoomDock } from "../../../features/session/components/RoomDock";
import { SessionSummaryModal } from "../../../features/session/components/SessionSummaryModal";
import { leaveRoom } from "../../../features/session/session.api";
import { buildSummary, workInProgress, type SessionSummary } from "../../../features/session/summary";
import { useActor } from "../../../features/session/useActor";
import { useRoom } from "../../../features/session/useRoom";
import { useRoomActions } from "../../../features/session/useRoomActions";
import { getTimer, type TimerState } from "../../../features/timer/timer.api";
import { TimerPanel } from "../../../features/timer/TimerPanel";

/**
 * One study room: the dock with everyone in it, the user's own timer, and
 * Leave (UC-03) with its confirmation and session summary.
 */
export default function RoomPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  const actorState = useActor();
  const actor = actorState.status === "ready" ? actorState.actor : null;
  const live = useRoom(actor, id);
  const actions = useRoomActions(actor, () => void live.refresh());
  const { room, participants, timers, presence } = live;

  // Leave flow: confirm -> leave -> summary -> back to the lake.
  const [confirming, setConfirming] = useState<{ timer: TimerState | null } | null>(null);
  const [leaving, setLeaving] = useState(false);
  const [leaveError, setLeaveError] = useState<string | null>(null);
  const [summary, setSummary] = useState<SessionSummary | null>(null);

  // Removed by someone else (kick, idle, timeout, room closed): back to the
  // list with the reason.
  useEffect(() => {
    if (presence.kind === "removed") {
      router.replace(`/rooms?notice=${presence.reason || "SESSION_ENDED"}`);
    } else if (presence.kind === "not_found") {
      router.replace("/rooms?notice=ROOM_NOT_FOUND");
    }
  }, [presence, router]);

  const me = participants.find((p) => p.user_id === actor?.userId);

  async function askToLeave() {
    if (!actor) return;
    setLeaveError(null);
    // Read our own timer fresh: the warning and the summary both need it.
    const timer = await getTimer({ sessionId: id, userId: actor.userId }).catch(() => timers[actor.userId] ?? null);
    setConfirming({ timer });
  }

  async function confirmLeave() {
    if (!actor || !confirming || !room) return;
    setLeaving(true);
    live.markLeaving();
    try {
      const result = await leaveRoom(actor, id);
      const built = await buildSummary({
        userId: actor.userId,
        roomName: room.name,
        joinedAt: result.joined_at || me?.joined_at || new Date().toISOString(),
        leftAt: result.left_at,
        timer: confirming.timer,
      });
      setConfirming(null);
      setSummary(built);
    } catch {
      setLeaveError("Couldn't leave the room. Check your connection and try again.");
    } finally {
      setLeaving(false);
    }
  }

  if (actorState.status === "signed_out") {
    return (
      <Centered>
        <PixelPanel className="text-center">
          <h1 className="font-display text-3xl leading-none">Sign in to fish here</h1>
          <p className="mt-3 text-sm text-bark">Rooms are for signed-in anglers.</p>
          <Link href="/signin" className="pixel-btn mt-5">
            Sign in
          </Link>
        </PixelPanel>
      </Centered>
    );
  }

  if (!room || presence.kind === "loading" || !actor) {
    return (
      <Centered wide>
        <div className="pixel-panel">
          <div className="pixel-skeleton h-10 w-1/2" />
          <div className="pixel-skeleton mt-6 h-44" />
        </div>
      </Centered>
    );
  }

  if (presence.kind === "visitor") {
    const full = room.participant_count >= room.participant_limit;
    return (
      <Centered>
        <PixelPanel className="text-center">
          <p className="font-label text-[10px] uppercase tracking-[0.12em] text-bark">Study room</p>
          <h1 className="font-display text-4xl leading-none">{room.name}</h1>
          <p className="mt-3 text-sm text-bark">
            {room.status === "ENDED"
              ? "This room has closed."
              : `${room.participant_count} of ${room.participant_limit} spots taken.`}
          </p>
          {actions.joinError && <PixelAlert className="mt-4 text-left">{actions.joinError}</PixelAlert>}
          <div className="mt-6 flex flex-col-reverse justify-center gap-3 sm:flex-row">
            <Link href="/rooms" className="pixel-btn pixel-btn--ghost">
              Back to the lake
            </Link>
            {room.status === "ACTIVE" && (
              <PixelButton onClick={() => void actions.join(id)} disabled={full || actions.joiningId !== null}>
                {full ? "Room is full" : actions.joiningId ? "Joining…" : "Join this room"}
              </PixelButton>
            )}
          </div>
        </PixelPanel>
        <AlreadyInRoomModal
          conflict={actions.conflict}
          resolving={actions.resolving}
          onLeaveAndContinue={() => void actions.leaveCurrentAndContinue()}
          onCancel={actions.dismissConflict}
        />
      </Centered>
    );
  }

  const others = participants.filter((p) => p.user_id !== actor.userId).length;

  return (
    <main className="flex flex-1 justify-center px-4 pt-6 pb-16 sm:pt-10">
      <div className="grid w-full max-w-5xl items-start gap-6 md:grid-cols-[minmax(0,1fr)_minmax(0,24rem)]">
        <PixelPanel as="section" aria-labelledby="room-heading" className="flex flex-col gap-5">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <Link href="/rooms" className="font-label text-[10px] uppercase tracking-[0.12em] text-bark hover:text-ink">
                ← The lake
              </Link>
              <h1 id="room-heading" className="mt-1 truncate font-display text-4xl leading-none" title={room.name}>
                {room.name}
              </h1>
              <p className="mt-2 text-sm text-bark">
                {room.status === "ENDED"
                  ? "This room has closed."
                  : room.participant_limit === 1
                  ? "A solo dock. Just you and the water."
                  : others === 0
                    ? `Waiting for anglers · ${room.participant_count}/${room.participant_limit} spots`
                    : `You and ${others} other${others > 1 ? "s" : ""} · ${room.participant_count}/${room.participant_limit} spots`}
              </p>
            </div>
            <PixelButton variant="danger" className="flex-none" onClick={() => void askToLeave()}>
              Leave
            </PixelButton>
          </div>

          {live.stale && (
            <PixelAlert tone="warn">Lost touch with the room. Showing the last known dock while we reconnect.</PixelAlert>
          )}

          <RoomDock participants={participants} limit={room.participant_limit} timers={timers} meId={actor.userId} />
        </PixelPanel>

        {/* The timer is what the user came for: first on phones. */}
        <div className="order-first md:order-none md:sticky md:top-24">
          <TimerPanel owner={{ sessionId: id, userId: actor.userId }} />
        </div>
      </div>

      <LeaveRoomDialog
        open={confirming !== null}
        roomName={room.name}
        workInProgress={workInProgress(confirming?.timer ?? null)}
        lastOne={participants.length <= 1}
        leaving={leaving}
        error={leaveError}
        onCancel={() => setConfirming(null)}
        onConfirm={() => void confirmLeave()}
      />
      <SessionSummaryModal summary={summary} onDone={() => router.push("/rooms")} />
    </main>
  );
}

function Centered({ children, wide = false }: { children: ReactNode; wide?: boolean }) {
  return (
    <main className="flex flex-1 items-start justify-center px-4 pt-12 pb-16">
      <div className={wide ? "w-full max-w-5xl" : "w-full max-w-md"}>{children}</div>
    </main>
  );
}

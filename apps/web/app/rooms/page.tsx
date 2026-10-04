"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelPanel } from "../../components/ui/PixelPanel";
import { AlreadyInRoomModal } from "../../features/session/components/AlreadyInRoomModal";
import { CreateRoomModal } from "../../features/session/components/CreateRoomModal";
import { RoomCard } from "../../features/session/components/RoomCard";
import { useActor } from "../../features/session/useActor";
import { useRoomActions } from "../../features/session/useRoomActions";
import { useRoomList } from "../../features/session/useRoomList";

/** Why the user was sent back here from a room (?notice=), UC-03 E-3/E-4/E-6/E-7. */
const NOTICES: Record<string, string> = {
  KICKED: "An admin removed you from that room.",
  IDLE_TIMEOUT: "No focus block ran for 10 minutes, so you were moved off the dock.",
  DISCONNECT_TIMEOUT: "We lost your connection, so you were moved off the dock.",
  SESSION_ENDED: "That room has closed.",
  ROOM_NOT_FOUND: "That room doesn't exist any more.",
};

/**
 * The lake: every active room (UC-02 step 2), refreshed live, plus opening a
 * new one (UC-01).
 */
export default function RoomsPage() {
  const router = useRouter();
  const actorState = useActor();
  const actor = actorState.status === "ready" ? actorState.actor : null;
  const { rooms, mine, error, refresh } = useRoomList(actor);
  const actions = useRoomActions(actor, () => void refresh());
  const [creating, setCreating] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    const code = new URLSearchParams(window.location.search).get("notice");
    if (code && NOTICES[code]) setNotice(NOTICES[code]);
  }, []);

  const myRoom = mine?.in_session ? mine.session : undefined;
  const openRooms = rooms?.filter((r) => r.participant_count < r.participant_limit).length ?? 0;

  return (
    <main className="flex flex-1 items-start justify-center px-4 pt-8 pb-16 sm:pt-12">
      <div className="w-full max-w-2xl">
        <PixelPanel as="section" aria-labelledby="rooms-heading">
          <div className="flex flex-wrap items-end justify-between gap-4">
            <div>
              <p className="font-label text-[10px] uppercase tracking-[0.12em] text-bark">Study rooms</p>
              <h1 id="rooms-heading" className="font-display text-4xl leading-none">
                The Lake
              </h1>
              <p className="mt-2 text-sm text-bark">Pick a spot on a dock, or open your own. Everyone runs their own timer.</p>
            </div>
            {actorState.status !== "signed_out" && (
              <PixelButton onClick={() => setCreating(true)} disabled={!actor}>
                + Open a room
              </PixelButton>
            )}
          </div>

          {notice && (
            <PixelAlert tone="warn" className="mt-5 flex items-start justify-between gap-3">
              <span>{notice}</span>
              <button type="button" className="pixel-link text-sm" onClick={() => setNotice(null)}>
                OK
              </button>
            </PixelAlert>
          )}

          {actorState.status === "signed_out" ? (
            <div className="pixel-placeholder mt-6 text-center">
              <p>Sign in to join a room and start fishing for focus.</p>
              <Link href="/signin" className="pixel-btn mt-4">
                Sign in
              </Link>
            </div>
          ) : (
            <>
              {myRoom && (
                <div className="pixel-tile pixel-room-card pixel-room-card--mine mt-6">
                  <div className="min-w-0 flex-1">
                    <p className="pixel-room-card__meta">Your line is still in the water</p>
                    <p className="pixel-room-card__name">{myRoom.name}</p>
                  </div>
                  <Link href={`/rooms/${myRoom.id}`} className="pixel-btn">
                    Return
                  </Link>
                </div>
              )}

              {actions.joinError && (
                <PixelAlert className="mt-5 flex items-start justify-between gap-3">
                  <span>{actions.joinError}</span>
                  <button type="button" className="pixel-link text-sm" onClick={actions.clearJoinError}>
                    OK
                  </button>
                </PixelAlert>
              )}

              <div className="mt-6 flex items-center justify-between gap-3">
                <h2 className="pixel-label m-0">
                  {rooms ? `${rooms.length} room${rooms.length === 1 ? "" : "s"} · ${openRooms} with space` : "Rooms"}
                </h2>
                {error && <span className="font-label text-[10px] uppercase text-rust">{error}</span>}
              </div>

              {rooms === null ? (
                <ul className="mt-3 flex flex-col gap-3" aria-busy="true">
                  {[0, 1, 2].map((i) => (
                    <li key={i} className="pixel-skeleton h-20" />
                  ))}
                </ul>
              ) : rooms.length === 0 ? (
                <div className="pixel-placeholder mt-3 flex flex-col items-center gap-4 py-10 text-center">
                  <p className="font-display text-2xl leading-none">The lake is quiet</p>
                  <p className="text-sm text-bark">No rooms are open. Be the first to cast a line.</p>
                  <PixelButton onClick={() => setCreating(true)} disabled={!actor}>
                    Open the first room
                  </PixelButton>
                </div>
              ) : (
                <ul className="mt-3 flex flex-col gap-3">
                  {[...rooms]
                    .sort((a, b) => Number(b.id === myRoom?.id) - Number(a.id === myRoom?.id))
                    .map((room) => (
                    <RoomCard
                      key={room.id}
                      room={room}
                      mine={room.id === myRoom?.id}
                      busy={actions.joiningId === room.id}
                      onJoin={() => void actions.join(room.id)}
                      onOpen={() => router.push(`/rooms/${room.id}`)}
                    />
                  ))}
                </ul>
              )}
            </>
          )}
        </PixelPanel>
      </div>

      <CreateRoomModal open={creating} onClose={() => setCreating(false)} onCreate={actions.create} />
      <AlreadyInRoomModal
        conflict={actions.conflict}
        resolving={actions.resolving}
        onLeaveAndContinue={() => void actions.leaveCurrentAndContinue()}
        onCancel={actions.dismissConflict}
      />
    </main>
  );
}

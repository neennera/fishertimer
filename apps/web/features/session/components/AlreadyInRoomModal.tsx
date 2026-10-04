"use client";

import Link from "next/link";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelModal } from "../../../components/ui/PixelModal";
import type { RoomConflict } from "../useRoomActions";

export interface AlreadyInRoomModalProps {
  conflict: RoomConflict | null;
  resolving: boolean;
  onLeaveAndContinue: () => void;
  onCancel: () => void;
}

/** UC-01 E-2 / UC-02 E-3: one room at a time. Leave the old one, or cancel. */
export function AlreadyInRoomModal({ conflict, resolving, onLeaveAndContinue, onCancel }: AlreadyInRoomModalProps) {
  if (!conflict) return null;
  const action = conflict.pending.kind === "join" ? "join this room" : "open a new room";

  return (
    <PixelModal open onClose={onCancel} title="You're already on a dock">
      <p>
        You&apos;re still in <strong className="font-display text-lg">{conflict.current.name}</strong>. Anglers fish
        one room at a time.
      </p>
      <p className="text-bark">
        Leave it to {action}. A focus block running there will be stopped and won&apos;t catch a fish.
      </p>
      <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:items-center sm:justify-between">
        <Link href={`/rooms/${conflict.current.id}`} className="pixel-link text-sm">
          Go back to {conflict.current.name}
        </Link>
        <div className="flex flex-col-reverse gap-3 sm:flex-row">
          <PixelButton variant="ghost" onClick={onCancel} disabled={resolving}>
            Cancel
          </PixelButton>
          <PixelButton variant="danger" onClick={onLeaveAndContinue} disabled={resolving}>
            {resolving ? "Leaving…" : "Leave it & continue"}
          </PixelButton>
        </div>
      </div>
    </PixelModal>
  );
}

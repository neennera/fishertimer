"use client";

import { PixelAlert } from "../../../components/ui/PixelAlert";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelModal } from "../../../components/ui/PixelModal";

export interface LeaveRoomDialogProps {
  open: boolean;
  roomName: string;
  /** A work cycle is running or paused: leaving forfeits it (UC-03 S-1). */
  workInProgress: boolean;
  /** The user is the last one here: leaving closes the room (UC-03 S-2). */
  lastOne: boolean;
  leaving: boolean;
  error: string | null;
  onCancel: () => void;
  onConfirm: () => void;
}

/** Confirms leaving. Cancel keeps the user in the room, timer untouched (E-1). */
export function LeaveRoomDialog({
  open,
  roomName,
  workInProgress,
  lastOne,
  leaving,
  error,
  onCancel,
  onConfirm,
}: LeaveRoomDialogProps) {
  return (
    <PixelModal open={open} onClose={onCancel} title="Reel in and leave?">
      <p>
        You&apos;re leaving <strong className="font-display text-lg">{roomName}</strong>. Fish you already
        caught here stay in your tank.
      </p>
      {workInProgress && (
        <PixelAlert tone="warn" className="mt-4">
          Your focus block is still in progress. It won&apos;t be completed and won&apos;t catch a fish if you
          leave now.
        </PixelAlert>
      )}
      {lastOne && <p className="mt-3 text-bark">You&apos;re the last one on the dock, so the room will close.</p>}
      {error && <PixelAlert className="mt-4">{error}</PixelAlert>}
      <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <PixelButton variant="ghost" onClick={onCancel} disabled={leaving}>
          Stay
        </PixelButton>
        <PixelButton variant="danger" onClick={onConfirm} disabled={leaving}>
          {leaving ? "Leaving…" : workInProgress ? "Leave anyway" : "Leave room"}
        </PixelButton>
      </div>
    </PixelModal>
  );
}

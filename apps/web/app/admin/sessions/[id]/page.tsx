'use client';

import { useCallback, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { PixelAlert } from '../../../../components/ui/PixelAlert';
import { PixelPanel } from '../../../../components/ui/PixelPanel';
import { PixelBadge } from '../../../../components/ui/PixelBadge';
import { PixelButton } from '../../../../components/ui/PixelButton';
import { PixelInput } from '../../../../components/ui/PixelInput';
import { PixelModal } from '../../../../components/ui/PixelModal';
import { endSession, getSession, kickParticipant } from '../../../../features/admin/admin.api';
import { SeatPips } from '../../../../features/admin/components/SeatPips';
import { formatDateTime } from '../../../../features/admin/components/format';
import {
  AdminError,
  StaleNotice,
  UNKNOWN_OUTCOME_MESSAGE,
  adminErrorMessage,
  isUnknownOutcome,
} from '../../../../features/admin/components/AdminError';
import { usePolling } from '../../../../features/admin/hooks/usePolling';
import { AdminApiError } from '../../../../lib/admin-api';
import type { AdminParticipant } from '../../../../features/admin/types';

/**
 * Cases where the page should refresh and say what happened rather than keep
 * the dialog open for a blind retry: a 409 (the room or member changed since it
 * was shown) or no answer in time (the action may have gone through).
 */
const isStateChanged = (e: unknown) => (e instanceof AdminApiError && e.status === 409) || isUnknownOutcome(e);
const changedMessage = (e: unknown, fallback: string) =>
  isUnknownOutcome(e) ? UNKNOWN_OUTCOME_MESSAGE : adminErrorMessage(e, fallback);

/** Admin — one live session's detail and its members, refreshed every few seconds. */
export default function AdminSessionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const fetchSession = useCallback(() => getSession(id), [id]);
  const { data, error, stale, refresh } = usePolling(fetchSession, id);
  /** getSession answers null once the room is no longer live (UC-04 E-3). */
  const session = data ?? null;
  const ended = data === null;
  const [notice, setNotice] = useState<string | null>(null);
  const [kickTarget, setKickTarget] = useState<AdminParticipant | null>(null);
  const [kickReason, setKickReason] = useState('');
  const [endReason, setEndReason] = useState('');
  const [kicking, setKicking] = useState(false);
  const [kickError, setKickError] = useState<string | null>(null);
  const [confirmingEnd, setConfirmingEnd] = useState(false);
  const [ending, setEnding] = useState(false);
  const [endError, setEndError] = useState<string | null>(null);

  function closeKick() {
    if (!kicking) {
      setKickTarget(null);
      setKickError(null);
      setKickReason('');
    }
  }

  async function confirmKick() {
    if (!kickTarget) return;
    setKicking(true);
    setKickError(null);
    try {
      const res = await kickParticipant(id, kickTarget.user_id, kickReason);
      const name = kickTarget.display_name ?? kickTarget.user_id;
      setNotice(res.session_ended ? `${name} was removed, and the room closed because it was empty.` : `${name} was removed from the room.`);
      setKickTarget(null);
      setKickReason('');
      await refresh();
    } catch (e: unknown) {
      if (isStateChanged(e)) {
        // Member already left or room ended: nothing changed, show the current state (UC-10 E-3).
        setNotice(changedMessage(e, 'This room has changed.'));
        setKickTarget(null);
        setKickReason('');
        await refresh();
        return;
      }
      setKickError(adminErrorMessage(e, 'Failed to remove the angler'));
    } finally {
      setKicking(false);
    }
  }

  function closeEnd() {
    if (!ending) {
      setConfirmingEnd(false);
      setEndError(null);
      setEndReason('');
    }
  }

  async function confirmEnd() {
    setEnding(true);
    setEndError(null);
    try {
      await endSession(id, endReason);
      setNotice('Room closed. Everyone was sent back to the lake.');
      setConfirmingEnd(false);
      setEndReason('');
      await refresh();
    } catch (e: unknown) {
      if (isStateChanged(e)) {
        setNotice(changedMessage(e, 'This room has already closed.'));
        setConfirmingEnd(false);
        setEndReason('');
        await refresh();
        return;
      }
      setEndError(adminErrorMessage(e, 'Failed to close the room'));
    } finally {
      setEnding(false);
    }
  }

  const name = (p: AdminParticipant | null) => p?.display_name ?? 'this angler';

  return (
    <main className="flex flex-1 items-start justify-center px-4 pt-8 pb-16 sm:pt-12">
      <div className="w-full max-w-2xl">
        <PixelPanel as="section" aria-labelledby="room-heading">
          <Link href="/admin" className="font-label text-[10px] uppercase tracking-[0.12em] text-bark hover:text-ink">
            ← Live rooms
          </Link>

          {notice && (
            <PixelAlert tone="warn" className="mt-4 flex items-start justify-between gap-3">
              <span>{notice}</span>
              <button type="button" className="pixel-link text-sm" onClick={() => setNotice(null)}>
                OK
              </button>
            </PixelAlert>
          )}
          {error ? <AdminError error={error} fallback="Failed to load the room" onRetry={refresh} /> : null}

          {ended && (
            <div className="pixel-placeholder mt-5 flex flex-col items-center gap-3 py-10 text-center">
              <p className="font-display text-2xl leading-none">This room has closed</p>
              <p className="text-sm text-bark">It is no longer on the lake.</p>
              <Link href="/admin" className="pixel-btn">
                Back to live rooms
              </Link>
            </div>
          )}

          {data === undefined && !error && (
            <div aria-busy="true">
              <div className="pixel-skeleton mt-3 h-10 w-1/2" />
              <div className="pixel-skeleton mt-6 h-44" />
            </div>
          )}

          {session && (
            <>
              <div className="mt-1 flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <h1 id="room-heading" className="truncate font-display text-4xl leading-none" title={session.title}>
                    {session.title}
                  </h1>
                  <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
                    <SeatPips taken={session.participant_count} limit={session.max_participants} />
                    <span className="pixel-room-card__meta">
                      {session.participant_count}/{session.max_participants} anglers · opened{' '}
                      {formatDateTime(session.created_at)}
                    </span>
                  </div>
                </div>
                <PixelButton variant="danger" className="flex-none" onClick={() => setConfirmingEnd(true)}>
                  Close room
                </PixelButton>
              </div>

              <div className="mt-6 flex items-center justify-between gap-3">
                <h2 className="pixel-label m-0">On the dock ({session.participants.length})</h2>
                <StaleNotice stale={stale} />
              </div>
              {session.participants.length === 0 ? (
                <p className="pixel-placeholder mt-3 py-6 text-center text-sm text-bark">Nobody is on the dock.</p>
              ) : (
                <ul className="mt-3 flex flex-col gap-3">
                  {session.participants.map((p) => (
                    <li key={p.user_id} className="pixel-tile pixel-room-card">
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <span className="truncate font-display text-xl leading-none">
                            {p.display_name ?? 'Unnamed angler'}
                          </span>
                          {p.is_host && <PixelBadge className="pixel-badge--wood">HOST</PixelBadge>}
                        </div>
                        <p className="pixel-room-card__meta mt-1.5">Joined {formatDateTime(p.joined_at)}</p>
                      </div>
                      {!p.is_host && (
                        <PixelButton
                          variant="danger"
                          className="pixel-btn--sm flex-none"
                          onClick={() => setKickTarget(p)}
                          aria-label={`Remove ${name(p)}`}
                        >
                          Remove
                        </PixelButton>
                      )}
                    </li>
                  ))}
                </ul>
              )}
            </>
          )}
        </PixelPanel>
      </div>

      <PixelModal open={kickTarget !== null} onClose={closeKick} title="Remove angler?">
        <p className="text-sm">
          Remove <strong>{name(kickTarget)}</strong> from this room? Their timer stops and any focus block in
          progress is lost.
          {session?.participant_count === 1 && <> They are the last angler, so the room will close too.</>}
        </p>
        <PixelInput
          label="Reason (optional)"
          className="mt-3"
          value={kickReason}
          onChange={(e) => setKickReason(e.target.value)}
          maxLength={200}
          disabled={kicking}
        />
        {kickError && <PixelAlert className="mt-3">{kickError}</PixelAlert>}
        <div className="mt-5 flex flex-col-reverse justify-end gap-3 sm:flex-row">
          <PixelButton variant="ghost" onClick={closeKick} disabled={kicking}>
            Cancel
          </PixelButton>
          <PixelButton variant="danger" onClick={confirmKick} disabled={kicking}>
            {kicking ? 'Removing…' : 'Remove'}
          </PixelButton>
        </div>
      </PixelModal>

      <PixelModal open={confirmingEnd} onClose={closeEnd} title="Close room?">
        <p className="text-sm">
          Close <strong>{session?.title}</strong>? Everyone on the dock is sent back to the lake, their focus blocks
          in progress are lost, and the room cannot be reopened.
        </p>
        <PixelInput
          label="Reason (optional)"
          className="mt-3"
          value={endReason}
          onChange={(e) => setEndReason(e.target.value)}
          maxLength={200}
          disabled={ending}
        />
        {endError && <PixelAlert className="mt-3">{endError}</PixelAlert>}
        <div className="mt-5 flex flex-col-reverse justify-end gap-3 sm:flex-row">
          <PixelButton variant="ghost" onClick={closeEnd} disabled={ending}>
            Cancel
          </PixelButton>
          <PixelButton variant="danger" onClick={confirmEnd} disabled={ending}>
            {ending ? 'Closing…' : 'Close room'}
          </PixelButton>
        </div>
      </PixelModal>
    </main>
  );
}

'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { SessionHeader } from '../../../../components/SessionHeader';
import { ParallaxScene } from '../../../../components/ui/ParallaxScene';
import { ADMIN_SCENE_LAYERS } from '../../../../lib/scenes/admin-scene';
import { PixelPanel } from '../../../../components/ui/PixelPanel';
import { PixelBadge } from '../../../../components/ui/PixelBadge';
import { PixelButton } from '../../../../components/ui/PixelButton';
import { PixelInput } from '../../../../components/ui/PixelInput';
import { PixelModal } from '../../../../components/ui/PixelModal';
import { endSession, getSession, kickParticipant } from '../../../../features/admin/admin.api';
import { AdminNav } from '../../../../features/admin/components/AdminNav';
import { StatusBadge } from '../../../../features/admin/components/StatusBadge';
import { formatDateTime } from '../../../../features/admin/components/format';
import type { AdminParticipant, AdminSessionDetail } from '../../../../features/admin/types';

/** Admin — one session's detail and its members. */
export default function AdminSessionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [session, setSession] = useState<AdminSessionDetail | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [kickTarget, setKickTarget] = useState<AdminParticipant | null>(null);
  const [kickReason, setKickReason] = useState('');
  const [endReason, setEndReason] = useState('');
  const [kicking, setKicking] = useState(false);
  const [kickError, setKickError] = useState<string | null>(null);
  const [confirmingEnd, setConfirmingEnd] = useState(false);
  const [ending, setEnding] = useState(false);
  const [endError, setEndError] = useState<string | null>(null);

  useEffect(() => {
    getSession(id)
      .then((s) => (s ? setSession(s) : setNotFound(true)))
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'Failed to load session'));
  }, [id]);

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
      const updated = await kickParticipant(id, kickTarget.user_id, kickReason);
      if (!updated) {
        setKickError('Member is no longer in this session.');
        return;
      }
      setSession(updated);
      setKickTarget(null);
      setKickReason('');
    } catch (e: unknown) {
      setKickError(e instanceof Error ? e.message : 'Failed to kick member');
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
      const updated = await endSession(id, endReason);
      if (!updated) {
        setEndError('This session has already ended.');
        return;
      }
      setSession(updated);
      setConfirmingEnd(false);
      setEndReason('');
    } catch (e: unknown) {
      setEndError(e instanceof Error ? e.message : 'Failed to close session');
    } finally {
      setEnding(false);
    }
  }

  return (
    <div className="min-h-screen">
      <ParallaxScene layers={ADMIN_SCENE_LAYERS} />
      <SessionHeader />
      <main className="mx-auto w-full max-w-3xl px-4 py-8 sm:py-10">
        <PixelPanel>
          <AdminNav active="sessions" />
          <Link href="/admin" className="text-bark hover:text-ink font-label text-sm">
            ← All sessions
          </Link>
          <div className="mb-5" />

          {error && (
            <div className="pixel-alert" role="alert">
              ⚠️ {error}
            </div>
          )}
          {notFound && <p className="font-body text-bark">Session not found.</p>}
          {!session && !error && !notFound && (
            <p className="font-body text-bark" aria-busy="true">
              Loading…
            </p>
          )}

          {session && (
            <>
              <div className="flex items-center gap-2 flex-wrap">
                <h1 className="font-display leading-tight" style={{ fontSize: 'calc(var(--px) * 9)' }}>
                  {session.title}
                </h1>
                {/* <StatusBadge active={session.is_active} /> */}
                {session.is_active && (
                  <PixelButton
                    variant="danger"
                    className="ml-auto"
                    onClick={() => setConfirmingEnd(true)}
                  >
                    Close session
                  </PixelButton>
                )}
              </div>

              <dl className="grid grid-cols-2 sm:grid-cols-4 gap-2.5 mt-5">
                {[
                  ['Host', session.host_name],
                  ['Members', `${session.participant_count}/${session.max_participants}`],
                  ['Started', formatDateTime(session.created_at)],
                  ['Ended', session.ended_at ? formatDateTime(session.ended_at) : '—'],
                ].map(([label, value]) => (
                  <div
                    key={label}
                    className="pixel-tile text-left"
                  >
                    <dt className="font-label text-[10px] uppercase text-bark">{label}</dt>
                    <dd className="font-numeric text-sm text-ink mt-0.5 truncate">{value}</dd>
                  </div>
                ))}
              </dl>

              <h2 className="font-label text-sm uppercase text-bark mt-6 mb-2">
                Members ({session.participants.length})
              </h2>
              <ul className="flex flex-col gap-2">
                {session.participants.map((p) => (
                  <li
                    key={p.user_id}
                    className="pixel-tile flex flex-wrap items-center justify-between gap-2 text-left"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="font-numeric text-sm text-ink truncate">{p.display_name}</span>
                      {p.is_host && <PixelBadge className="pixel-badge--wood">HOST</PixelBadge>}
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="font-label text-[11px] text-bark">
                        Joined {formatDateTime(p.joined_at)} ·{' '}
                        {p.left_at ? `Left ${formatDateTime(p.left_at)}` : 'In room'}
                      </span>
                      {session.is_active && !p.left_at && !p.is_host && (
                        <PixelButton
                          variant="danger"
                          className="pixel-btn--sm"
                          onClick={() => setKickTarget(p)}
                          aria-label={`Kick ${p.display_name}`}
                        >
                          Kick
                        </PixelButton>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            </>
          )}
        </PixelPanel>
      </main>

      <PixelModal open={kickTarget !== null} onClose={closeKick} title="Kick member?">
        <p className="font-body text-sm">
          Remove <strong>{kickTarget?.display_name}</strong> from this session? They will be
          disconnected from the room immediately.
          {session?.participant_count === 1 && (
            <> They are the last member, so the session will also be closed.</>
          )}
        </p>
        <PixelInput
          label="Reason (optional)"
          className="mt-3"
          value={kickReason}
          onChange={(e) => setKickReason(e.target.value)}
          maxLength={200}
          disabled={kicking}
        />
        {kickError && (
          <div className="pixel-alert mt-3" role="alert">
            ⚠️ {kickError}
          </div>
        )}
        <div className="mt-4 flex justify-end gap-2">
          <PixelButton variant="ghost" onClick={closeKick} disabled={kicking}>
            Cancel
          </PixelButton>
          <PixelButton variant="danger" onClick={confirmKick} disabled={kicking}>
            {kicking ? 'Kicking…' : 'Kick'}
          </PixelButton>
        </div>
      </PixelModal>

      <PixelModal open={confirmingEnd} onClose={closeEnd} title="Close session?">
        <p className="font-body text-sm">
          Close <strong>{session?.title}</strong>? Everyone still in the room will be
          disconnected and the session cannot be reopened.
        </p>
        <PixelInput
          label="Reason (optional)"
          className="mt-3"
          value={endReason}
          onChange={(e) => setEndReason(e.target.value)}
          maxLength={200}
          disabled={ending}
        />
        {endError && (
          <div className="pixel-alert mt-3" role="alert">
            ⚠️ {endError}
          </div>
        )}
        <div className="mt-4 flex justify-end gap-2">
          <PixelButton variant="ghost" onClick={closeEnd} disabled={ending}>
            Cancel
          </PixelButton>
          <PixelButton variant="danger" onClick={confirmEnd} disabled={ending}>
            {ending ? 'Closing…' : 'Close session'}
          </PixelButton>
        </div>
      </PixelModal>
    </div>
  );
}

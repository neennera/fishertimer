'use client';

import Link from 'next/link';
import { cx } from '../../lib/cx';
import { PixelPanel } from '../../components/ui/PixelPanel';
import { listSessions } from '../../features/admin/admin.api';
import { AdminNav } from '../../features/admin/components/AdminNav';
import { AdminError, StaleNotice } from '../../features/admin/components/AdminError';
import { SeatPips } from '../../features/admin/components/SeatPips';
import { formatDateTime } from '../../features/admin/components/format';
import { usePolling } from '../../features/admin/hooks/usePolling';

/** Admin — every live room (refreshed every few seconds), each opening its detail + members. */
export default function AdminPage() {
  const { data: sessions, error, stale } = usePolling(listSessions, 'sessions');
  const full = sessions?.filter((s) => s.participant_count >= s.max_participants).length ?? 0;

  return (
    <main className="flex flex-1 items-start justify-center px-4 pt-8 pb-16 sm:pt-12">
      <div className="w-full max-w-2xl">
        <PixelPanel as="section" aria-labelledby="admin-heading">
          <div className="flex flex-wrap items-end justify-between gap-4">
            <div>
              <p className="font-label text-[10px] uppercase tracking-[0.12em] text-bark">Admin</p>
              <h1 id="admin-heading" className="font-display text-4xl leading-none">
                Live Rooms
              </h1>
              <p className="mt-2 text-sm text-bark">
                Every open room on the lake. Inspect one to remove an angler or close it.
              </p>
            </div>
            <AdminNav active="sessions" />
          </div>

          {error ? <AdminError error={error} fallback="Failed to load rooms" /> : null}

          {!error && (
            <>
              <div className="mt-6 flex items-center justify-between gap-3">
                <h2 className="pixel-label m-0">
                  {sessions
                    ? `${sessions.length} room${sessions.length === 1 ? '' : 's'} live · ${full} full`
                    : 'Rooms'}
                </h2>
                <StaleNotice stale={stale} />
              </div>

              {!sessions ? (
                <ul className="mt-3 flex flex-col gap-3" aria-busy="true">
                  {[0, 1, 2].map((i) => (
                    <li key={i} className="pixel-skeleton h-20" />
                  ))}
                </ul>
              ) : sessions.length === 0 ? (
                <div className="pixel-placeholder mt-3 flex flex-col items-center gap-2 py-10 text-center">
                  <p className="font-display text-2xl leading-none">The lake is quiet</p>
                  <p className="text-sm text-bark">No rooms are open right now.</p>
                </div>
              ) : (
                <ul className="mt-3 flex flex-col gap-3">
                  {sessions.map((s) => (
                    <li
                      key={s.session_id}
                      className={cx(
                        'pixel-tile pixel-room-card',
                        s.participant_count >= s.max_participants && 'pixel-room-card--full',
                      )}
                    >
                      <div className="min-w-0 flex-1">
                        <p className="pixel-room-card__name" title={s.title}>
                          {s.title}
                        </p>
                        <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
                          <SeatPips taken={s.participant_count} limit={s.max_participants} />
                          <span className="pixel-room-card__meta">
                            {s.participant_count}/{s.max_participants} anglers · opened{' '}
                            {formatDateTime(s.created_at)}
                          </span>
                        </div>
                      </div>
                      <Link href={`/admin/sessions/${s.session_id}`} className="pixel-btn">
                        Inspect
                      </Link>
                    </li>
                  ))}
                </ul>
              )}
            </>
          )}
        </PixelPanel>
      </div>
    </main>
  );
}

'use client';

import { useCallback, useState } from 'react';
import { cx } from '../../../lib/cx';
import { PixelPanel } from '../../../components/ui/PixelPanel';
import { PixelBadge } from '../../../components/ui/PixelBadge';
import { PixelButton } from '../../../components/ui/PixelButton';
import { listLogs } from '../../../features/admin/admin.api';
import { AdminNav } from '../../../features/admin/components/AdminNav';
import { AdminError, StaleNotice } from '../../../features/admin/components/AdminError';
import { formatDateTime } from '../../../features/admin/components/format';
import { usePolling } from '../../../features/admin/hooks/usePolling';
import type { AdminLogAction } from '../../../features/admin/types';

/** The service returns at most this many entries (services/admin DefaultLogLimit). */
const LOG_LIMIT = 100;

const FILTERS: { value: AdminLogAction | undefined; label: string }[] = [
  { value: undefined, label: 'All' },
  { value: 'KICK_USER', label: 'Removals' },
  { value: 'FORCE_CLOSE_SESSION', label: 'Closures' },
];

const ACTION_LABEL: Record<AdminLogAction, string> = {
  KICK_USER: 'Angler removed',
  FORCE_CLOSE_SESSION: 'Room closed',
};

/** admin_logs stores IDs only; show the first block, full ID on hover. */
const shortId = (id: string) => id.split('-')[0];

/** Admin — audit trail of every moderation action (admin_db.admin_logs), newest first. */
export default function AdminLogsPage() {
  const [filter, setFilter] = useState<AdminLogAction | undefined>(undefined);
  const fetchLogs = useCallback(() => listLogs(filter), [filter]);
  // Refreshes like the room pages; changing the filter restarts with a fresh load.
  const { data: logs, error, stale, refresh } = usePolling(fetchLogs, filter ?? 'all');

  return (
    <main className="flex flex-1 items-start justify-center px-4 pt-8 pb-16 sm:pt-12">
      <div className="w-full max-w-2xl">
        <PixelPanel as="section" aria-labelledby="logs-heading">
          <div className="flex flex-wrap items-end justify-between gap-4">
            <div>
              <p className="font-label text-[10px] uppercase tracking-[0.12em] text-bark">Admin</p>
              <h1 id="logs-heading" className="font-display text-4xl leading-none">
                Audit Log
              </h1>
              <p className="mt-2 text-sm text-bark">The latest anglers removed and rooms closed by admins.</p>
            </div>
            <AdminNav active="logs" />
          </div>

          {error ? <AdminError error={error} fallback="Failed to load the audit log" onRetry={refresh} /> : null}

          {!error && (
            <>
              <div className="mt-6 flex flex-wrap items-center justify-between gap-3">
                <h2 className="pixel-label m-0">
                  {logs
                    ? logs.length >= LOG_LIMIT
                      ? `Latest ${LOG_LIMIT} entries`
                      : `${logs.length} entr${logs.length === 1 ? 'y' : 'ies'}`
                    : 'Entries'}
                </h2>
                <StaleNotice stale={stale} />
                <div className="flex flex-wrap gap-2" role="group" aria-label="Filter by action">
                  {FILTERS.map((f) => (
                    <PixelButton
                      key={f.label}
                      variant={f.value === filter ? 'primary' : 'ghost'}
                      className="pixel-btn--sm"
                      aria-pressed={f.value === filter}
                      onClick={() => setFilter(f.value)}
                    >
                      {f.label}
                    </PixelButton>
                  ))}
                </div>
              </div>

              {!logs ? (
                <ul className="mt-3 flex flex-col gap-3" aria-busy="true">
                  {[0, 1, 2].map((i) => (
                    <li key={i} className="pixel-skeleton h-16" />
                  ))}
                </ul>
              ) : logs.length === 0 ? (
                <div className="pixel-placeholder mt-3 py-10 text-center">
                  <p className="font-display text-2xl leading-none">All calm</p>
                  <p className="mt-2 text-sm text-bark">No moderation actions yet.</p>
                </div>
              ) : (
                <ul className="mt-3 flex flex-col gap-3">
                  {logs.map((l) => (
                    <li
                      key={l.log_id}
                      className={cx(
                        'pixel-tile pixel-room-card',
                        l.action === 'FORCE_CLOSE_SESSION' && 'pixel-room-card--full',
                      )}
                    >
                      <div className="min-w-0 flex-1">
                        <div className="flex flex-wrap items-center justify-between gap-2">
                          <PixelBadge className={l.action === 'KICK_USER' ? undefined : 'pixel-badge--wood'}>
                            {ACTION_LABEL[l.action]}
                          </PixelBadge>
                          <time className="pixel-room-card__meta" dateTime={l.created_at}>
                            {formatDateTime(l.created_at)}
                          </time>
                        </div>
                        <p className="mt-2 text-sm">{l.reason ?? <span className="text-bark">No reason given</span>}</p>
                        <p className="pixel-room-card__meta mt-1.5">
                          <span title={l.target_id}>
                            {l.action === 'KICK_USER' ? 'Angler' : 'Room'} {shortId(l.target_id)}
                          </span>{' '}
                          · by admin <span title={l.admin_id}>{shortId(l.admin_id)}</span>
                        </p>
                      </div>
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

'use client';

import { useEffect, useState } from 'react';
import { SessionHeader } from '../../../components/SessionHeader';
import { ParallaxScene } from '../../../components/ui/ParallaxScene';
import { ADMIN_SCENE_LAYERS } from '../../../lib/scenes/admin-scene';
import { PixelPanel } from '../../../components/ui/PixelPanel';
import { PixelBadge } from '../../../components/ui/PixelBadge';
import { PixelButton } from '../../../components/ui/PixelButton';
import { listLogs } from '../../../features/admin/admin.api';
import { AdminNav } from '../../../features/admin/components/AdminNav';
import { formatDateTime } from '../../../features/admin/components/format';
import type { AdminLog, AdminLogAction } from '../../../features/admin/types';

const FILTERS: { value: AdminLogAction | undefined; label: string }[] = [
  { value: undefined, label: 'All' },
  { value: 'KICK_USER', label: 'Kicks' },
  { value: 'FORCE_CLOSE_SESSION', label: 'Closures' },
];

const ACTION_LABEL: Record<AdminLogAction, string> = {
  KICK_USER: 'KICK USER',
  FORCE_CLOSE_SESSION: 'FORCE CLOSE',
};

/** Admin — audit trail of every moderation action (admin_db.admin_logs). */
export default function AdminLogsPage() {
  const [filter, setFilter] = useState<AdminLogAction | undefined>(undefined);
  const [logs, setLogs] = useState<AdminLog[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let stale = false;
    listLogs(filter)
      .then((l) => !stale && setLogs(l))
      .catch((e: unknown) => !stale && setError(e instanceof Error ? e.message : 'Failed to load logs'));
    return () => {
      stale = true;
    };
  }, [filter]);

  return (
    <div className="min-h-screen">
      <ParallaxScene layers={ADMIN_SCENE_LAYERS} />
      <SessionHeader />
      <main className="mx-auto w-full max-w-3xl px-4 py-8 sm:py-10">
        <PixelPanel>
          <AdminNav active="logs" />

          <h1 className="font-display leading-tight" style={{ fontSize: 'calc(var(--px) * 10)' }}>
            Admin — Audit Log
          </h1>
          <p className="mt-1 text-bark font-body text-sm">
            {logs ? `${logs.length} entries ` : 'Loading…'}
            {/* {logs ? `${logs.length} entries, newest first` : 'Loading…'} */}
          </p>

          <div className="mt-4 flex flex-wrap gap-2" role="group" aria-label="Filter by action">
            {FILTERS.map((f) => (
              <PixelButton
                key={f.label}
                variant={f.value === filter ? 'primary' : 'ghost'}
                className="pixel-btn--sm"
                aria-pressed={f.value === filter}
                onClick={() => {
                  setLogs(null);
                  setFilter(f.value);
                }}
              >
                {f.label}
              </PixelButton>
            ))}
          </div>

          {error && (
            <div className="pixel-alert mt-4" role="alert">
              ⚠️ {error}
            </div>
          )}

          <ul className="mt-4 flex flex-col gap-2" aria-busy={!logs}>
            {!logs &&
              !error &&
              [1, 2, 3].map((i) => (
                <li key={i} className="pixel-tile animate-pulse" style={{ height: '4rem' }} aria-hidden="true" />
              ))}

            {logs?.map((l) => (
              <li key={l.log_id} className="pixel-tile text-left">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-2 min-w-0">
                    <PixelBadge className={l.action === 'KICK_USER' ? undefined : 'pixel-badge--wood'}>
                      {ACTION_LABEL[l.action]}
                    </PixelBadge>
                    <span className="font-numeric text-sm text-ink truncate">{l.target_label}</span>
                  </div>
                  <time className="font-label text-[11px] text-bark" dateTime={l.created_at}>
                    {formatDateTime(l.created_at)}
                  </time>
                </div>
                <div className="font-label text-[11px] text-bark mt-1.5">
                  By {l.admin_name} · {l.reason ? `Reason: ${l.reason}` : 'No reason given'}
                </div>
              </li>
            ))}

            {logs?.length === 0 && (
              <li className="text-bark font-body text-sm">No log entries.</li>
            )}
          </ul>
        </PixelPanel>
      </main>
    </div>
  );
}

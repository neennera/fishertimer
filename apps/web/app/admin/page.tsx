'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Header } from '../../components/Header';
import { ParallaxScene } from '../../components/ui/ParallaxScene';
import { ADMIN_SCENE_LAYERS } from '../../lib/scenes/admin-scene';
import { PixelPanel } from '../../components/ui/PixelPanel';
import { PixelButton } from '../../components/ui/PixelButton';
import { listSessions } from '../../features/admin/admin.api';
import { StatusBadge } from '../../features/admin/components/StatusBadge';
import { formatDateTime } from '../../features/admin/components/format';
import type { AdminSession } from '../../features/admin/types';
import { SessionHeader } from "../../components/SessionHeader";

/** Admin — list study sessions, with a link to each session's detail + members. */
export default function AdminPage() {
  const [sessions, setSessions] = useState<AdminSession[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listSessions()
      .then(setSessions)
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'Failed to load sessions'));
  }, []);

  return (
    <div className="min-h-screen">
      <ParallaxScene layers={ADMIN_SCENE_LAYERS} />
      <SessionHeader />
      <main className="mx-auto w-full max-w-3xl px-4 py-8 sm:py-10">
        <PixelPanel>
          {/* <div className="mb-5 pb-3 border-b border-[var(--color-rule)]">
            <Link href="/" className="text-bark hover:text-ink font-label text-sm">
              ← Back to Study Room
            </Link>
          </div> */}

          <h1 className="font-display leading-tight" style={{ fontSize: 'calc(var(--px) * 10)' }}>
            Admin — Sessions
          </h1>
          <p className="mt-1 text-bark font-body text-sm">
            {sessions ? `${sessions.length} active sessions` : 'Loading…'}
          </p>

          {error && (
            <div className="pixel-alert mt-4" role="alert">
              ⚠️ {error}
            </div>
          )}

          <ul className="mt-5 flex flex-col gap-2.5" aria-busy={!sessions}>
            {!sessions &&
              !error &&
              [1, 2, 3].map((i) => (
                <li
                  key={i}
                  className="pixel-tile animate-pulse"
                  style={{ height: '4.5rem' }}
                  aria-hidden="true"
                />
              ))}

            {sessions?.map((s) => (
              <li
                key={s.session_id}
                className="pixel-tile flex flex-wrap items-center justify-between gap-3 text-left"
              >
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-numeric text-base text-ink truncate">{s.title}</span>
                    <StatusBadge active={s.is_active} />
                  </div>
                  <div className="font-label text-[11px] text-bark mt-1">
                    Host {s.host_name} · 👥 {s.participant_count}/{s.max_participants} · Started{' '}
                    {formatDateTime(s.created_at)}
                  </div>
                </div>
                <Link href={`/admin/sessions/${s.session_id}`}>
                  <PixelButton variant="ghost">View detail →</PixelButton>
                </Link>
              </li>
            ))}

            {sessions?.length === 0 && (
              <li className="text-bark font-body text-sm">No active sessions right now.</li>
            )}
          </ul>
        </PixelPanel>
      </main>
    </div>
  );
}

import Link from 'next/link';

const LINKS = [
  { href: '/admin', label: '🛡️ Sessions', key: 'sessions' },
  { href: '/admin/logs', label: '📜 Audit Log', key: 'logs' },
] as const;

/** Tab links shared by every admin page. */
export function AdminNav({ active }: { active: 'sessions' | 'logs' }) {
  return (
    <nav
      aria-label="Admin"
      className="mb-5 flex flex-wrap items-center justify-end gap-2 pb-3 border-b border-[var(--color-rule)]"
    >
      {LINKS.map((l) => (
        <Link
          key={l.key}
          href={l.href}
          aria-current={l.key === active ? 'page' : undefined}
          className={
            l.key === active
              ? 'pixel-btn pixel-btn--sm'
              : 'pixel-btn pixel-btn--sm pixel-btn--ghost'
          }
        >
          {l.label}
        </Link>
      ))}
    </nav>
  );
}

import Link from 'next/link';

const LINKS = [
  { href: '/admin', label: 'Live rooms', key: 'sessions' },
  { href: '/admin/logs', label: 'Audit log', key: 'logs' },
] as const;

/** Tab links shared by every admin page. */
export function AdminNav({ active }: { active: 'sessions' | 'logs' }) {
  return (
    <nav aria-label="Admin" className="flex flex-wrap gap-2">
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

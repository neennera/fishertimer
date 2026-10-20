import Link from 'next/link';

const LINKS = [
  { href: '/admin', label: 'Live rooms', key: 'sessions' },
  { href: '/admin/logs', label: 'Audit log', key: 'logs' },
] as const;

/** Tab links shared by every admin page. */
export function AdminNav({ active, className }: { active: 'sessions' | 'logs'; className?: string }) {
  return (
    <nav aria-label="Admin" className={`flex flex-wrap gap-2 ${className ?? ''}`}>
      {LINKS.map((l) => (
        <Link
          key={l.key}
          href={l.href}
          aria-current={l.key === active ? 'page' : undefined}
          className={
            l.key === active
              ? 'pixel-btn pixel-btn--sm min-h-11 sm:min-h-0'
              : 'pixel-btn pixel-btn--sm pixel-btn--ghost min-h-11 sm:min-h-0'
          }
        >
          {l.label}
        </Link>
      ))}
    </nav>
  );
}

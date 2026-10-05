'use client';

import type { LeaderboardPeriod } from '../types';
import { cx } from '../../../lib/cx';

const PERIODS: { value: LeaderboardPeriod; label: string }[] = [
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
  { value: 'all-time', label: 'All-Time' },
];

export interface PeriodTabsProps {
  active: LeaderboardPeriod;
  onChange: (period: LeaderboardPeriod) => void;
  disabled?: boolean;
}

/**
 * PeriodTabs — S-2 filter strip (UC-08).
 *
 * Compact pixel-button tabs (Weekly, Monthly, All-Time) meant to share a row
 * with the refresh button. The active tab uses the primary variant.
 */
export function PeriodTabs({ active, onChange, disabled }: PeriodTabsProps) {
  return (
    <div role="tablist" aria-label="Ranking period" className="flex items-center gap-1.5">
      {PERIODS.map(({ value, label }) => (
        <button
          key={value}
          role="tab"
          aria-selected={active === value}
          disabled={disabled}
          onClick={() => onChange(value)}
          className={cx('pixel-btn transition-all', active === value ? '' : 'pixel-btn--ghost')}
          style={{ fontSize: 'calc(var(--px) * 4)', padding: '0.25rem 0.6rem', minHeight: 0 }}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

'use client';

import type { LeaderboardPeriod } from '../types';
import { cx } from '../../../lib/cx';

const PERIODS: { value: LeaderboardPeriod; label: string }[] = [
  { value: 'weekly', label: '📅 Weekly' },
  { value: 'monthly', label: '🗓️ Monthly' },
  { value: 'all-time', label: '🏆 All-Time' },
];

export interface PeriodTabsProps {
  active: LeaderboardPeriod;
  onChange: (period: LeaderboardPeriod) => void;
  disabled?: boolean;
}

/**
 * PeriodTabs — S-2 filter strip (UC-08).
 *
 * Renders three pixel-button tabs: Weekly, Monthly, All-Time.
 * The active tab uses the primary variant; others use ghost.
 * Switching tab calls `onChange` which triggers S-2 re-render in the parent.
 */
export function PeriodTabs({ active, onChange, disabled }: PeriodTabsProps) {
  return (
    <div
      role="tablist"
      aria-label="Ranking period"
      className="flex flex-wrap gap-2"
    >
      {PERIODS.map(({ value, label }) => (
        <button
          key={value}
          role="tab"
          aria-selected={active === value}
          disabled={disabled}
          onClick={() => onChange(value)}
          className={cx(
            'pixel-btn',
            active === value ? '' : 'pixel-btn--ghost',
          )}
          style={{ fontSize: 'calc(var(--px) * 5.5)' }}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

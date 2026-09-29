'use client';

import type { LeaderboardPeriod } from '../types';
import { cx } from '../../../lib/cx';

const PERIODS: { value: LeaderboardPeriod; label: string; icon: string; subtitle: string }[] = [
  { value: 'weekly', label: 'Weekly', icon: '📅', subtitle: 'Past 7 Days' },
  { value: 'monthly', label: 'Monthly', icon: '🗓️', subtitle: 'Past 30 Days' },
  { value: 'all-time', label: 'All-Time', icon: '🏆', subtitle: 'Hall of Fame' },
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
 * The active tab uses the primary amber variant; others use ghost.
 * Switching tab calls `onChange` which triggers S-2 re-render in the parent.
 */
export function PeriodTabs({ active, onChange, disabled }: PeriodTabsProps) {
  return (
    <div
      role="tablist"
      aria-label="Ranking period"
      className="grid grid-cols-3 gap-2 sm:gap-3"
    >
      {PERIODS.map(({ value, label, icon, subtitle }) => {
        const isSelected = active === value;
        return (
          <button
            key={value}
            role="tab"
            aria-selected={isSelected}
            disabled={disabled}
            onClick={() => onChange(value)}
            className={cx(
              'pixel-btn flex flex-col items-center justify-center text-center py-2.5 px-2 transition-all',
              isSelected ? '' : 'pixel-btn--ghost',
            )}
            style={{
              minHeight: '4rem',
            }}
          >
            <div className="flex items-center gap-1.5 font-numeric text-sm sm:text-base leading-none">
              <span className="text-base sm:text-lg" aria-hidden="true">
                {icon}
              </span>
              <span>{label}</span>
            </div>
            <span
              className="text-[10px] sm:text-xs font-label mt-1 opacity-80"
              style={{
                color: isSelected ? '#ffffff' : 'var(--color-bark)',
              }}
            >
              {subtitle}
            </span>
          </button>
        );
      })}
    </div>
  );
}

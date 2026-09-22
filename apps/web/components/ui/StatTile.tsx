import { cx } from "../../lib/cx";

export interface StatTileProps {
  /** Preformatted for display — "18h 40m", not 1120. */
  value: string;
  caption: string;
  className?: string;
}

/**
 * One figure from the UC-07 dashboard: sessions joined, total focus time,
 * rewards earned. Read-only by definition.
 */
export function StatTile({ value, caption, className }: StatTileProps) {
  return (
    <div className={cx("pixel-tile", className)}>
      <div className="pixel-tile__value">{value}</div>
      <div className="pixel-tile__caption">{caption}</div>
    </div>
  );
}

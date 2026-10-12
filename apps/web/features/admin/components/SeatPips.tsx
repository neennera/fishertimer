import { cx } from '../../../lib/cx';

/** Seats taken out of the room limit — same pips as the room list. */
export function SeatPips({ taken, limit }: { taken: number; limit: number }) {
  return (
    <span className="pixel-seats" role="img" aria-label={`${taken} of ${limit} spots taken`}>
      {Array.from({ length: limit }, (_, i) => (
        <span key={i} className={cx('pixel-seats__seat', i < taken && 'pixel-seats__seat--taken')} />
      ))}
    </span>
  );
}

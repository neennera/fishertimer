"use client";

import { FishArt } from "../../../components/account/FishTank";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelModal } from "../../../components/ui/PixelModal";
import { StatTile } from "../../../components/ui/StatTile";
import { fishSprite } from "../../../lib/fish-sprites";
import { formatFocusMinutes } from "../../../lib/format";
import type { SessionSummary } from "../summary";

export interface SessionSummaryModalProps {
  summary: SessionSummary | null;
  onDone: () => void;
}

/** UC-03 step 5: what the user got done in the room they just left. */
export function SessionSummaryModal({ summary, onDone }: SessionSummaryModalProps) {
  if (!summary) return null;
  const caught = summary.fish?.reduce((n, f) => n + f.count, 0) ?? 0;

  return (
    <PixelModal open onClose={onDone} title="Your catch">
      <p className="text-bark">
        You spent {formatFocusMinutes(summary.stayMinutes)} at <span className="text-ink">{summary.roomName}</span>.
      </p>

      <div className="mt-5 grid grid-cols-3 gap-2">
        <StatTile value={String(summary.completedCycles)} caption="focus blocks" />
        <StatTile value={formatFocusMinutes(summary.focusMinutes)} caption="focus time" />
        <StatTile value={summary.fish ? String(caught) : "—"} caption="fish caught" />
      </div>

      <div className="mt-6">
        <p className="pixel-label">Fish from this session</p>
        {summary.fish === null ? (
          <p className="mt-2 text-sm text-bark">The reward shop is unreachable right now. Your fish are safe in your tank.</p>
        ) : summary.fish.length === 0 ? (
          <p className="mt-2 text-sm text-bark">
            No bites this time. Finish a focus block of 15 minutes or more to catch a fish.
          </p>
        ) : (
          <ul className="pixel-haul mt-2">
            {summary.fish.map((fish) => (
              <li key={fish.name} className="pixel-tile pixel-fish-card">
                <FishArt sprite={fishSprite(fish)} still />
                <span className="pixel-fish-card__text">
                  <span className="pixel-fish-card__name">{fish.name}</span>
                  <span className="pixel-fish-card__count">
                    ×{fish.count} · {fish.rarity.toLowerCase()}
                  </span>
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>

      {summary.forfeitedCycle && (
        <p className="mt-4 text-sm text-bark">The focus block you left mid-way didn&apos;t count.</p>
      )}

      <div className="mt-6 flex justify-end">
        <PixelButton onClick={onDone}>Back to the lake</PixelButton>
      </div>
    </PixelModal>
  );
}

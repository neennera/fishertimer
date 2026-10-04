"use client";

import { useEffect, useState } from "react";
import { FishArt } from "../../components/account/FishTank";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelModal } from "../../components/ui/PixelModal";
import { fishSprite } from "../../lib/fish-sprites";
import { fishCaughtBetween, type CaughtFish } from "../session/summary";
import { MIN_REWARD_MINUTES } from "./timer.api";
import type { CompletedBlock } from "./useStudyTimer";

/** Reward may take a moment (or a retry) to land: look a few times. */
const LOOK_AFTER_MS = [800, 2000, 4000, 7000, 11000];

export interface CatchRevealProps {
  block: CompletedBlock | null;
  userId: string;
  onClose: () => void;
}

/**
 * UC-09 step 9: after a work block completes, reel in what it caught. The
 * fish are read from Reward (awarded at or after the block's end).
 */
export function CatchReveal({ block, userId, onClose }: CatchRevealProps) {
  const [fish, setFish] = useState<CaughtFish[] | null>(null);
  const [looking, setLooking] = useState(true);

  useEffect(() => {
    if (!block) return;
    let cancelled = false;
    const timers: number[] = [];
    const from = new Date(new Date(block.completedAt).getTime() - 5_000);

    setFish(null);
    setLooking(true);
    LOOK_AFTER_MS.forEach((delay, i) => {
      timers.push(
        window.setTimeout(async () => {
          const found = await fishCaughtBetween(userId, from, new Date(Date.now() + 60_000));
          if (cancelled) return;
          if (found && found.length > 0) {
            setFish(found);
            setLooking(false);
            timers.forEach((t) => window.clearTimeout(t));
          } else if (i === LOOK_AFTER_MS.length - 1) {
            setFish(found ?? []);
            setLooking(false);
          }
        }, delay),
      );
    });
    return () => {
      cancelled = true;
      timers.forEach((t) => window.clearTimeout(t));
    };
  }, [block, userId]);

  if (!block) return null;
  const tooShort = block.minutes > 0 && block.minutes < MIN_REWARD_MINUTES;

  return (
    <PixelModal open onClose={onClose} title="Focus block complete!">
      <p>
        {block.minutes > 0
          ? `${block.minutes} minute${block.minutes === 1 ? "" : "s"} of focus, done.`
          : "Focus block done."}{" "}
        Take a break or keep the streak going.
      </p>

      <div className="pixel-catch mt-5" aria-live="polite">
        {looking ? (
          <p className="pixel-catch__waiting">Reeling it in…</p>
        ) : fish && fish.length > 0 ? (
          <ul className="pixel-haul">
            {fish.map((f) => (
              <li key={f.name} className="pixel-tile pixel-fish-card">
                <FishArt sprite={fishSprite(f)} still />
                <span className="pixel-fish-card__text">
                  <span className="pixel-fish-card__name">{f.name}</span>
                  <span className="pixel-fish-card__count">
                    ×{f.count} · {f.rarity.toLowerCase()}
                  </span>
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-bark">
            {tooShort
              ? `No bite this time: blocks under ${MIN_REWARD_MINUTES} minutes don't catch fish.`
              : "Your catch is on its way to your tank. Check your account page in a moment."}
          </p>
        )}
      </div>

      <div className="mt-6 flex justify-end">
        <PixelButton onClick={onClose}>Nice!</PixelButton>
      </div>
    </PixelModal>
  );
}

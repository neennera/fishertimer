"use client";

import { useEffect, useState } from "react";
import { FishArt } from "../../components/account/FishTank";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelModal } from "../../components/ui/PixelModal";
import { cx } from "../../lib/cx";
import { fishSprite } from "../../lib/fish-sprites";
import { fishCaughtBetween, type CaughtFish } from "../session/summary";
import { MIN_REWARD_MINUTES } from "./timer.api";
import type { CompletedBlock } from "./useStudyTimer";

/** Reward may take a moment (or a retry) to land: look a few times. */
const LOOK_AFTER_MS = [800, 2000, 4000, 7000, 11000];

/** Shown when Reward sends a catch without its catalogue details. */
const UNKNOWN_FISH = "Mystery catch";

export interface CatchRevealProps {
  block: CompletedBlock | null;
  userId: string;
  onClose: () => void;
  /** Offered while the timer waits for a break: start it or skip it from here. */
  breakMinutes?: number;
  onStartBreak?: () => void;
  onSkipBreak?: () => void;
}

/**
 * UC-09 step 9: after a work block completes, reel in what it caught. The
 * fish are read from Reward (awarded at or after the block's end).
 */
export function CatchReveal({ block, userId, onClose, breakMinutes, onStartBreak, onSkipBreak }: CatchRevealProps) {
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
  const best = fish?.[0] ?? null;
  const caught = fish?.reduce((n, f) => n + f.count, 0) ?? 0;
  const stage = looking ? "waiting" : best ? "caught" : "empty";

  function then(action?: () => void) {
    onClose();
    action?.();
  }

  return (
    <PixelModal open onClose={onClose} title="Focus block complete!">
      <div className="pixel-catch-stage" data-stage={stage} aria-hidden="true">
        <span className="pixel-catch-stage__line" />
        <span className="pixel-catch-stage__bobber" />
        {best && (
          <span className="pixel-catch-stage__fish">
            <FishArt sprite={fishSprite(best)} still />
          </span>
        )}
        <span className="pixel-catch-stage__splash" />
      </div>

      {/* Tall enough for a caught fish, so the dialog doesn't jump when it lands. */}
      <div className="mt-4 flex min-h-28 flex-col items-center justify-center text-center" aria-live="polite">
        {looking ? (
          <p className="pixel-catch__waiting">Reeling it in…</p>
        ) : best ? (
          <>
            <p className="font-label text-[10px] uppercase tracking-[0.12em] text-bark">You caught</p>
            <p className="font-display text-4xl leading-none">{best.name || UNKNOWN_FISH}</p>
            <p className="mt-2 flex items-center justify-center gap-2">
              {best.rarity && (
                <span className={cx("pixel-chip", "pixel-rarity")} data-rarity={best.rarity}>
                  {best.rarity.toLowerCase()}
                </span>
              )}
              {best.count > 1 && <span className="font-numeric text-xl leading-none">×{best.count}</span>}
            </p>
            {fish && fish.length > 1 && (
              <p className="mt-2 text-xs text-bark">
                and{" "}
                {fish
                  .slice(1)
                  .map((f) => `${f.name || UNKNOWN_FISH}${f.count > 1 ? ` ×${f.count}` : ""}`)
                  .join(", ")}
              </p>
            )}
          </>
        ) : (
          <p className="text-sm text-bark">
            {tooShort
              ? `No bite this time: blocks under ${MIN_REWARD_MINUTES} minutes don't catch fish.`
              : "Your catch is on its way to your tank. Check your account page in a moment."}
          </p>
        )}
      </div>

      <dl className="pixel-catch-stats mt-5">
        <div className="pixel-catch-stats__item">
          <dt>Focus</dt>
          <dd>
            {block.minutes > 0 ? block.minutes : "—"}
            <small>min</small>
          </dd>
        </div>
        <div className="pixel-catch-stats__item">
          <dt>Catch</dt>
          <dd>{looking ? "…" : caught}</dd>
        </div>
      </dl>

      <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        {onStartBreak ? (
          <>
            <PixelButton variant="ghost" onClick={() => then(onSkipBreak)}>
              Skip break
            </PixelButton>
            <PixelButton onClick={() => then(onStartBreak)}>
              Start {breakMinutes ? `${breakMinutes} min ` : ""}break
            </PixelButton>
          </>
        ) : (
          <PixelButton onClick={onClose}>Nice!</PixelButton>
        )}
      </div>
    </PixelModal>
  );
}

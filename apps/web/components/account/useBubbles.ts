"use client";

import { useCallback, useEffect, useRef, useState, type MouseEvent, type RefObject } from "react";

// Sprite sizes in art pixels (drawn at --px, like the fish).
const SIZES = { big: 8, small: 6 } as const;
export type BubbleSize = keyof typeof SIZES;

// Ambient: every 10–30s, a short stream of 1–3 bubbles from one spot.
const SPAWN_MIN_S = 10;
const SPAWN_MAX_S = 30;
const STREAM_MAX = 3;
const STREAM_GAP_MS: [number, number] = [250, 550];
// Easter egg: after this many quick taps on the tank, each further quick
// tap may release a bubble where it landed.
const TAP_STREAK = 5;
const TAP_GAP_MS = 700;
const TAP_BUBBLE_CHANCE = 0.35;
const MAX_BUBBLES = 8;
// Screen px per second.
const RISE_SPEED: [number, number] = [30, 44];
export const BUBBLE_IN_S = 0.3;
export const BUBBLE_POP_S = 0.5;
// A bubble fades after rising this long, or at the surface if sooner.
const BUBBLE_LIFE_S = 3;

export interface Bubble {
  id: number;
  size: BubbleSize;
  /** In --px units, from the water's left and bottom edges. */
  left: number;
  bottom: number;
  /** In --px units: the climb, which ends with the sprite's top at the surface. */
  rise: number;
  riseSeconds: number;
  /** Seconds after spawning that it starts to fade. */
  fadeAt: number;
  swaySeconds: number;
}

function between(min: number, max: number) {
  return min + Math.random() * (max - min);
}

/**
 * Occasional bubbles in the fish tank, plus a tap easter egg. Off under
 * prefers-reduced-motion and while the tab is hidden. Returns the live
 * bubbles, a remover for when a bubble's pop ends, and the tank's click
 * handler.
 */
export function useBubbles(waterRef: RefObject<HTMLElement | null>) {
  const [bubbles, setBubbles] = useState<Bubble[]>([]);
  const nextId = useRef(0);
  const taps = useRef({ count: 0, last: 0 });

  /** `at` is in screen px from the water's left / bottom edges. */
  const spawn = useCallback(
    (at?: { left: number; bottom: number }) => {
      const water = waterRef.current;
      if (!water || window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
        return;
      }
      const artPx = parseFloat(getComputedStyle(water).getPropertyValue("--px")) || 1;
      const width = water.clientWidth / artPx;
      const height = water.clientHeight / artPx;
      const size: BubbleSize = Math.random() < 0.6 ? "small" : "big";
      const box = SIZES[size];
      if (width < box || height < box) {
        return;
      }

      const centre = at ? at.left / artPx : between(0, width);
      const left = Math.min(Math.max(0, centre - box / 2), width - box);
      // Ambient bubbles start on the sand; tapped ones where the tap landed.
      const base = at ? at.bottom / artPx - box / 2 : between(0.04, 0.14) * height;
      const bottom = Math.min(Math.max(0, base), height - box);
      const rise = height - box - bottom;
      const riseSeconds = Math.max(0.2, (rise * artPx) / between(...RISE_SPEED));

      setBubbles((current) =>
        current.length >= MAX_BUBBLES
          ? current
          : [
              ...current,
              {
                id: nextId.current++,
                size,
                left,
                bottom,
                rise,
                riseSeconds,
                fadeAt: BUBBLE_IN_S + Math.min(riseSeconds, BUBBLE_LIFE_S),
                swaySeconds: between(1.1, 1.7),
              },
            ],
      );
    },
    [waterRef],
  );

  const remove = useCallback((id: number) => {
    setBubbles((current) => current.filter((bubble) => bubble.id !== id));
  }, []);

  // Ambient streams, rescheduled after each one.
  useEffect(() => {
    const timers = new Set<ReturnType<typeof setTimeout>>();
    const later = (fn: () => void, ms: number) => {
      const t = setTimeout(() => {
        timers.delete(t);
        fn();
      }, ms);
      timers.add(t);
    };
    function schedule() {
      later(() => {
        const water = waterRef.current;
        if (water && !document.hidden) {
          // One spot for the whole stream.
          const at = { left: between(0.05, 0.95) * water.clientWidth, bottom: between(0.06, 0.14) * water.clientHeight };
          const count = 1 + Math.floor(Math.random() * STREAM_MAX);
          let delay = 0;
          for (let i = 0; i < count; i++) {
            later(() => spawn(at), delay);
            delay += between(...STREAM_GAP_MS);
          }
        }
        schedule();
      }, between(SPAWN_MIN_S, SPAWN_MAX_S) * 1000);
    }
    schedule();
    return () => timers.forEach(clearTimeout);
  }, [spawn, waterRef]);

  const onTankClick = useCallback(
    (event: MouseEvent<HTMLElement>) => {
      const now = performance.now();
      const streak = taps.current;
      streak.count = now - streak.last <= TAP_GAP_MS ? streak.count + 1 : 1;
      streak.last = now;
      if (streak.count >= TAP_STREAK && Math.random() < TAP_BUBBLE_CHANCE) {
        const rect = event.currentTarget.getBoundingClientRect();
        spawn({ left: event.clientX - rect.left, bottom: rect.bottom - event.clientY });
      }
    },
    [spawn],
  );

  return { bubbles, remove, onTankClick };
}

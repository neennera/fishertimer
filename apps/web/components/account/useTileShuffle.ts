"use client";

import { useCallback, useLayoutEffect, useRef, type RefObject } from "react";

// Each tile gets its own start and length, picked fresh per sort, so they
// leave and land out of step rather than as one block.
const SHUFFLE_MS: [number, number] = [360, 540];
const DELAY_MS: [number, number] = [0, 140];
const EASING = "cubic-bezier(0.2, 0.8, 0.2, 1)";

function between([min, max]: [number, number]) {
  return min + Math.random() * (max - min);
}

/**
 * Slides tiles from their old grid positions to their new ones when the list
 * is re-sorted (FLIP). Call `snapshot()` just before changing the order; the
 * move plays once React has laid out the new order. Tiles are found by
 * `[data-tile]` (their key), so they must keep their React keys across sorts.
 */
export function useTileShuffle(listRef: RefObject<HTMLElement | null>, order: unknown) {
  // Where each tile was, relative to the list, when the sort was clicked.
  const before = useRef<Map<string, DOMRect> | null>(null);

  const snapshot = useCallback(() => {
    const list = listRef.current;
    if (!list) {
      return;
    }
    const origin = list.getBoundingClientRect();
    const rects = new Map<string, DOMRect>();
    // Includes any move still running, so a second click carries on from
    // where the tiles are on screen.
    for (const el of list.querySelectorAll<HTMLElement>("[data-tile]")) {
      const r = el.getBoundingClientRect();
      rects.set(el.dataset.tile!, new DOMRect(r.x - origin.x, r.y - origin.y));
    }
    before.current = rects;
  }, [listRef]);

  useLayoutEffect(() => {
    const list = listRef.current;
    const rects = before.current;
    before.current = null;
    if (!list || !rects || window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      return;
    }
    const tiles = Array.from(list.querySelectorAll<HTMLElement>("[data-tile]"));
    tiles.forEach((el) => el.getAnimations().forEach((a) => a.cancel()));
    const origin = list.getBoundingClientRect();
    tiles.forEach((el) => {
      const from = rects.get(el.dataset.tile!);
      if (!from) {
        return;
      }
      const to = el.getBoundingClientRect();
      const dx = from.x - (to.x - origin.x);
      const dy = from.y - (to.y - origin.y);
      if (dx === 0 && dy === 0) {
        return;
      }
      el.animate(
        [{ transform: `translate(${dx}px, ${dy}px)` }, { transform: "none" }],
        { duration: between(SHUFFLE_MS), delay: between(DELAY_MS), easing: EASING, fill: "backwards" },
      );
    });
  }, [listRef, order]);

  return snapshot;
}

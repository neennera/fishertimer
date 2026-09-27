"use client";

import { useMemo, useRef, type CSSProperties, type ReactNode } from "react";
import type { RewardItem, UserReward } from "@fishertimer/shared-types";
import { cx } from "../../lib/cx";
import { fishSprite } from "../../lib/fish-sprites";
import { FISHTANK_LAYERS } from "../../lib/scenes/fishtank";
import { useFishSwim, type SwimStart } from "./useFishSwim";

export interface FishTankProps {
  catalog: RewardItem[];
  /** Several rows for one species = its count. */
  rewards: UserReward[];
  emptyMessage: ReactNode;
}

interface CaughtSpecies {
  item: RewardItem;
  count: number;
}

// Phones show only the first 8 (pixel.css).
const MAX_TANK_FISH = 15;

// Halton points: deterministic (matches the server render) and evenly spread
// for any prefix, since phones show only the first 8 fish.
function halton(index: number, base: number) {
  let result = 0;
  let fraction = 1 / base;
  for (let i = index; i > 0; i = Math.floor(i / base)) {
    result += (i % base) * fraction;
    fraction /= base;
  }
  return result;
}

function restSpot(index: number): SwimStart {
  return { restX: halton(index + 1, 2), restY: halton(index + 1, 3) };
}

// One fish per catch, taken one per species per round, so every species
// shows before any repeats.
function tankFish(species: CaughtSpecies[]): { key: string; itemId: string }[] {
  const fish: { key: string; itemId: string }[] = [];
  for (let round = 0; fish.length < MAX_TANK_FISH; round++) {
    const before = fish.length;
    for (const { item, count } of species) {
      if (round < count && fish.length < MAX_TANK_FISH) {
        fish.push({ key: `${item.id}-${round}`, itemId: item.id });
      }
    }
    if (fish.length === before) {
      break;
    }
  }
  return fish;
}

function caughtSpecies(catalog: RewardItem[], rewards: UserReward[]): CaughtSpecies[] {
  const counts = new Map<string, number>();
  for (const reward of rewards) {
    counts.set(reward.itemId, (counts.get(reward.itemId) ?? 0) + 1);
  }

  return catalog
    .filter((item) => item.itemType === "FISH_SPECIES" && counts.has(item.id))
    .map((item) => ({ item, count: counts.get(item.id)! }));
}

export function countFishCaught(catalog: RewardItem[], rewards: UserReward[]) {
  return caughtSpecies(catalog, rewards).reduce((sum, { count }) => sum + count, 0);
}

function FishArt({ itemId, still = false }: { itemId: string; still?: boolean }) {
  const sprite = fishSprite(itemId);
  const strip = sprite.src !== undefined && (sprite.frames ?? 1) > 1;
  const style = (
    sprite.src
      ? { backgroundImage: `url("${sprite.src}")` }
      : { "--fish-color": `var(--color-${sprite.placeholder})` }
  ) as CSSProperties;
  // data-frame: tail pose, stepped by useFishSwim.
  const animated = !still && (strip || !sprite.src);

  return (
    <span
      className={cx(
        "pixel-fish-art",
        sprite.src ? "pixel-fish-art--sprite" : "pixel-fish-art--placeholder",
        strip && "pixel-fish-art--strip",
      )}
    >
      <span
        className="pixel-fish-art__body"
        style={style}
        data-frame={animated ? "0" : undefined}
      />
    </span>
  );
}

export function FishTank({ catalog, rewards, emptyMessage }: FishTankProps) {
  const species = useMemo(() => caughtSpecies(catalog, rewards), [catalog, rewards]);
  const swimmers = useMemo(() => tankFish(species), [species]);
  const starts = useMemo(() => swimmers.map((_, index) => restSpot(index)), [swimmers]);
  const waterRef = useRef<HTMLDivElement>(null);
  useFishSwim(waterRef, starts);

  return (
    <div className="flex flex-col gap-6">
      <div className="pixel-tank" aria-hidden="true">
        <div ref={waterRef} className="pixel-tank__water">
          {FISHTANK_LAYERS.map((layer) => (
            <div
              key={layer.src}
              className={cx("pixel-tank__layer", layer.front && "pixel-tank__layer--front")}
              style={
                {
                  backgroundImage: `url("${layer.src}")`,
                  "--layer-w": layer.width,
                  "--layer-offset": layer.offset ?? 0,
                } as CSSProperties
              }
            />
          ))}
          {swimmers.map(({ key, itemId }, index) => {
            const start = starts[index]!;
            const style = {
              "--fish-rest-x": start.restX,
              "--fish-rest-y": start.restY,
            } as CSSProperties;

            return (
              <span key={key} className="pixel-fish" style={style} data-fish="">
                <span className="pixel-fish__turn">
                  <FishArt itemId={itemId} />
                </span>
              </span>
            );
          })}
        </div>
      </div>

      {species.length === 0 ? (
        <p className="text-center text-sm text-bark">{emptyMessage}</p>
      ) : (
        <ul className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3">
          {species.map(({ item, count }) => (
            <li key={item.id} className="pixel-tile pixel-fish-card">
              <FishArt itemId={item.id} still />
              <span className="pixel-fish-card__text">
                <span className="pixel-fish-card__name">{item.itemName}</span>
                <span className="pixel-fish-card__count">×{count}</span>
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

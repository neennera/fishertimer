"use client";

import {
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type ReactNode,
} from "react";
import { cx } from "../../lib/cx";
import { fishSprite, type FishSprite } from "../../lib/fish-sprites";
import type { RewardRarity, RewardSummaryItem } from "../../lib/profile-types";
import { FISHTANK_LAYERS } from "../../lib/scenes/fishtank";
import { BUBBLE_IN_S, BUBBLE_POP_S, useBubbles } from "./useBubbles";
import { useFishSwim, type SwimStart } from "./useFishSwim";
import { useTileShuffle } from "./useTileShuffle";

export interface FishTankProps {
  /** Only FISH items go in the tank. */
  items: RewardSummaryItem[];
  emptyMessage: ReactNode;
}

interface CaughtSpecies {
  // Rewards have no item id, so the list position is the key.
  key: string;
  name: string;
  count: number;
  rarity: RewardRarity;
  sprite: FishSprite;
}

// Lowest first. Points mirror ScoreForRarity in the reward service
// (services/reward/internal/domain/entity.go): keep the two in step.
// tankWeight: how much more likely a catch of that rarity is to swim.
const RARITIES: {
  rarity: RewardRarity;
  label: string;
  points: number;
  tankWeight: number;
}[] = [
  { rarity: "COMMON", label: "Common", points: 10, tankWeight: 1 },
  { rarity: "UNCOMMON", label: "Uncommon", points: 25, tankWeight: 2 },
  { rarity: "RARE", label: "Rare", points: 50, tankWeight: 3 },
  { rarity: "EPIC", label: "Epic", points: 100, tankWeight: 5 },
  { rarity: "LEGENDARY", label: "Legendary", points: 250, tankWeight: 8 },
];

function rarityRank(rarity: RewardRarity) {
  const rank = RARITIES.findIndex((r) => r.rarity === rarity);
  return rank === -1 ? 0 : rank;
}

function rarityLabel(rarity: RewardRarity) {
  return RARITIES[rarityRank(rarity)]!.label;
}

type SortKey = "recent" | "count" | "rarity";

const SORTS: { key: SortKey; label: string }[] = [
  // The order the API sends: most recently caught species first.
  { key: "recent", label: "Recent" },
  { key: "count", label: "Count" },
  { key: "rarity", label: "Rarity" },
];

// Array.sort is stable, so ties keep the "recent" order.
function sortSpecies(species: CaughtSpecies[], sort: SortKey): CaughtSpecies[] {
  if (sort === "count") {
    return [...species].sort((a, b) => b.count - a.count);
  }
  if (sort === "rarity") {
    return [...species].sort(
      (a, b) => rarityRank(b.rarity) - rarityRank(a.rarity),
    );
  }
  return species;
}

// Seeded, so the pick is stable across re-renders.
function mulberry32(seed: number) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// Phones show only the first 8 (pixel.css).
const MAX_TANK_FISH = 15;

interface TankFish {
  key: string;
  rarity: RewardRarity;
  sprite: FishSprite;
}

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

// A weighted random draw, without replacement, from every catch: rarer fish
// are more likely to swim, and no species shows more often than it was
// caught. Sorted by draw key, so the first 8 (phones) are weighted too.
function tankFish(species: CaughtSpecies[], seed: number): TankFish[] {
  const random = mulberry32(seed);
  const drawn: (TankFish & { draw: number })[] = [];
  for (const { key, count, rarity, sprite } of species) {
    const weight = RARITIES[rarityRank(rarity)]!.tankWeight;
    // More than MAX_TANK_FISH of one species can never all be picked.
    for (let n = 0; n < Math.min(count, MAX_TANK_FISH); n++) {
      // Efraimidis-Spirakis: the largest random^(1/weight) keys win.
      drawn.push({
        key: `${key}-${n}`,
        rarity,
        sprite,
        draw: random() ** (1 / weight),
      });
    }
  }
  return drawn
    .sort((a, b) => b.draw - a.draw)
    .slice(0, MAX_TANK_FISH)
    .map(({ key, rarity, sprite }) => ({ key, rarity, sprite }));
}

function caughtSpecies(items: RewardSummaryItem[]): CaughtSpecies[] {
  return items
    .filter((item) => item.type === "FISH" && item.count > 0)
    .map((item, index) => ({
      key: `${index}-${item.name}`,
      name: item.name,
      count: item.count,
      rarity: item.rarity,
      sprite: fishSprite(item),
    }));
}

export function countFishCaught(items: RewardSummaryItem[]) {
  return caughtSpecies(items).reduce((sum, { count }) => sum + count, 0);
}

function FishArt({
  sprite,
  rarity,
  still = false,
}: {
  sprite: FishSprite;
  rarity: RewardRarity;
  still?: boolean;
}) {
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
      data-rarity={rarity}
    >
      <span
        className="pixel-fish-art__body"
        style={style}
        data-frame={animated ? "0" : undefined}
      />
      {rarity === "EPIC" && (
        // Clipped to the fish: the sprite is its own mask.
        <span
          className="pixel-fish-art__glare"
          style={
            sprite.src
              ? ({ "--fish-mask": `url("${sprite.src}")` } as CSSProperties)
              : undefined
          }
        />
      )}
      {rarity === "LEGENDARY" && (
        <span className="pixel-fish-art__sparkles">
          <i />
          <i />
          <i />
        </span>
      )}
    </span>
  );
}

export function FishTank({ items, emptyMessage }: FishTankProps) {
  const species = useMemo(() => caughtSpecies(items), [items]);
  // One draw per page load: sorting the tiles never reshuffles the tank.
  const [seed] = useState(() => Math.floor(Math.random() * 2 ** 32));
  const swimmers = useMemo(() => tankFish(species, seed), [species, seed]);
  const [sort, setSort] = useState<SortKey>("recent");
  const tiles = useMemo(() => sortSpecies(species, sort), [species, sort]);
  const tilesRef = useRef<HTMLUListElement>(null);
  const shuffleTiles = useTileShuffle(tilesRef, sort);
  const starts = useMemo(
    () => swimmers.map((_, index) => restSpot(index)),
    [swimmers],
  );
  const waterRef = useRef<HTMLDivElement>(null);
  const feed = useFishSwim(waterRef, starts);
  const { bubbles, remove: removeBubble, onTankClick } = useBubbles(waterRef);

  return (
    <div className="flex flex-col gap-6">
      <div className="pixel-tank" aria-hidden="true">
        {/* Clicks feed the bubble easter egg (useBubbles). */}
        <div ref={waterRef} className="pixel-tank__water" onClick={onTankClick}>
          {FISHTANK_LAYERS.map((layer) => (
            <div
              key={layer.src}
              className={cx(
                "pixel-tank__layer",
                layer.front && "pixel-tank__layer--front",
              )}
              style={
                {
                  backgroundImage: `url("${layer.src}")`,
                  "--layer-w": layer.width,
                  "--layer-offset": layer.offset ?? 0,
                } as CSSProperties
              }
            />
          ))}
          {swimmers.map(({ key, rarity, sprite }, index) => {
            const start = starts[index]!;
            const style = {
              "--fish-rest-x": start.restX,
              "--fish-rest-y": start.restY,
            } as CSSProperties;

            return (
              <span key={key} className="pixel-fish" style={style} data-fish="">
                <span className="pixel-fish__turn">
                  <FishArt sprite={sprite} rarity={rarity} />
                </span>
              </span>
            );
          })}
          {bubbles.map((bubble) => (
            <span
              key={bubble.id}
              className={cx(
                "pixel-tank-bubble",
                `pixel-tank-bubble--${bubble.size}`,
              )}
              style={
                {
                  "--bubble-left": bubble.left,
                  "--bubble-bottom": bubble.bottom,
                  "--bubble-rise": bubble.rise,
                  "--bubble-in-s": `${BUBBLE_IN_S}s`,
                  "--bubble-rise-s": `${bubble.riseSeconds}s`,
                  "--bubble-sway-s": `${bubble.swaySeconds}s`,
                  "--bubble-pop-at": `${bubble.fadeAt}s`,
                  "--bubble-pop-s": `${BUBBLE_POP_S}s`,
                } as CSSProperties
              }
            >
              <span className="pixel-tank-bubble__rise">
                <span className="pixel-tank-bubble__sway">
                  <span
                    className="pixel-tank-bubble__sprite"
                    onAnimationEnd={(event) => {
                      if (event.animationName === "pixel-bubble-pop") {
                        removeBubble(bubble.id);
                      }
                    }}
                  />
                </span>
              </span>
            </span>
          ))}
        </div>
      </div>

      {species.length === 0 ? (
        <p className="text-center text-sm text-bark">{emptyMessage}</p>
      ) : (
        <>
          <div className="pixel-fish-controls">
            <div
              className="pixel-fish-sort"
              role="group"
              aria-label="Sort fish"
            >
              <span className="pixel-fish-sort__label" aria-hidden="true">
                Sort
              </span>
              {SORTS.map(({ key, label }) => (
                <button
                  key={key}
                  type="button"
                  className={cx(
                    "pixel-btn pixel-btn--sm",
                    sort !== key && "pixel-btn--ghost",
                  )}
                  aria-pressed={sort === key}
                  onClick={() => {
                    if (key !== sort) {
                      shuffleTiles();
                      setSort(key);
                    }
                  }}
                >
                  {label}
                </button>
              ))}
            </div>
            <button
              type="button"
              className="pixel-btn pixel-btn--sm pixel-btn--ghost pixel-fish-feed"
              onClick={feed}
            >
              Feed
            </button>
          </div>
          <ul
            ref={tilesRef}
            className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3"
          >
            {tiles.map(({ key, name, count, rarity, sprite }) => (
              <li
                key={key}
                className="pixel-tile pixel-fish-card"
                data-rarity={rarity}
                data-tile={key}
              >
                <FishArt sprite={sprite} rarity={rarity} still />
                <span className="pixel-fish-card__text">
                  <span className="pixel-fish-card__name">{name}</span>
                  <span className="pixel-fish-card__count">×{count}</span>
                </span>
                <span className="pixel-fish-card__ribbon">
                  {rarityLabel(rarity)}
                </span>
              </li>
            ))}
          </ul>
          <ul className="pixel-rarity-legend" aria-label="Points per rarity">
            {RARITIES.map(({ rarity, label, points }) => (
              <li key={rarity} data-rarity={rarity}>
                <span className="pixel-rarity-legend__dot" aria-hidden="true" />
                <span>{label}</span>
                <span className="pixel-rarity-legend__points">
                  {points} pts
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
}

import { CSSProperties } from "react";
import { cx } from "../../lib/cx";

export interface ParallaxLayer {
  /** Path to a tiling sprite, e.g. from public/sprites/scene/. */
  src: string;
  /** Relative scroll speed. Higher = faster (nearer); lower = slower (farther). */
  speed: number;
}

export interface ParallaxSceneProps {
  layers: ParallaxLayer[];
  className?: string;
}

// Duration for a layer at speed = 1. Other layers scale inversely, so a
// slower (smaller) speed gets a longer duration, per DESIGN_SYSTEM.md's
// prefers-reduced-motion rule this stays a plain animation-duration — the
// global stylesheet's !important override in globals.css still wins.
const PARALLAX_BASE_DURATION_S = 60;

/**
 * A stack of tiling background sprites that scroll at different speeds.
 * Generic over its layer list — a signin scene, a room scene, whatever
 * later reuses this — so it takes layers as a prop rather than baking in
 * one scene's file names.
 */
export function ParallaxScene({ layers, className }: ParallaxSceneProps) {
  return (
    <div className={cx("pixel-parallax", className)}>
      {layers.map((layer) => {
        const style: CSSProperties = {
          backgroundImage: `url(${layer.src})`,
          animationDuration: `${PARALLAX_BASE_DURATION_S / layer.speed}s`,
        };

        return (
          <div key={layer.src} className="pixel-parallax__layer" style={style} />
        );
      })}
    </div>
  );
}

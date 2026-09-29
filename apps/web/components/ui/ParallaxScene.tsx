import { CSSProperties } from "react";
import { cx } from "../../lib/cx";

export type ParallaxLayer =
  | {
      src: string;
      /** Smooth painterly/gradient art: background-size: cover, no tiling. */
      fit: "cover";
      /** Seconds for one drift leg (there, then back — see pixel.css's
       * pixel-parallax-drift). Omit for a static layer, e.g. the sky itself. */
      driftSeconds?: number;
    }
  | {
      src: string;
      /** Pixel-art sprite: tiled horizontally, scaled by whole --px steps. */
      fit: "tile";
      /** Higher = faster (nearer); lower = slower (farther). */
      speed: number;
    };

export interface ParallaxSceneProps {
  layers: ParallaxLayer[];
  className?: string;
}

// Duration for a "tile" layer at speed = 1. Other layers scale inversely, so
// a slower (smaller) speed gets a longer duration, per DESIGN_SYSTEM.md's
// prefers-reduced-motion rule this stays a plain animation-duration — the
// global stylesheet's !important override in globals.css still wins.
const PARALLAX_BASE_DURATION_S = 60;

/**
 * A full-viewport stack of background layers, back to front. Generic over
 * its layer list — a signin scene, a room scene, whatever later reuses this
 * — so it takes layers as a prop rather than baking in one scene's files.
 *
 * Two kinds of layer, because one CSS technique can't serve both art styles
 * without either stretching or gapping (see pixel.css for the full case):
 * - "cover": smooth gradient/painterly art (the sky, clouds) with no pixel
 *   grid to misalign — background-size: cover always preserves aspect ratio
 *   and fully covers any viewport shape, boundless. Optionally drifts: since
 *   cover already fixes the image's scale, animating background-position-x
 *   as a percentage only pans within the art, it never rescales it — safe
 *   in a way a percentage-driven size never was for the tile sprites.
 * - "tile": pixel-art sprites (mountains, forest, water) scaled by whole
 *   multiples of --px and tiled horizontally, anchored to the bottom — a
 *   bounded band, not the whole viewport, so it never needs to stretch.
 *
 * A "cover" sky layer naturally fills whatever space opens up above the
 * "tile" band on a tall viewport, so the two compose without a gap.
 * Fixed to the viewport by its own CSS; no client-side sizing logic.
 */
export function ParallaxScene({ layers, className }: ParallaxSceneProps) {
  return (
    <div className={cx("pixel-parallax", className)} aria-hidden="true">
      {layers.map((layer) => {
        const style: CSSProperties = {
          backgroundImage: `url(${layer.src})`,
        };

        const drifts = layer.fit === "cover" && layer.driftSeconds !== undefined;

        if (layer.fit === "tile") {
          style.animationDuration = `${PARALLAX_BASE_DURATION_S / layer.speed}s`;
        } else if (drifts) {
          style.animationDuration = `${layer.driftSeconds}s`;
        }

        return (
          <div
            key={layer.src}
            className={cx(
              "pixel-parallax__layer",
              layer.fit === "cover" ? "pixel-parallax__layer--cover" : "pixel-parallax__layer--tile",
              drifts && "pixel-parallax__layer--drift",
            )}
            style={style}
          />
        );
      })}
    </div>
  );
}

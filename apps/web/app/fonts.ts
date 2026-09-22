import { DotGothic16, Jersey_15, Jersey_25, Silkscreen } from "next/font/google";

/**
 * Four faces, each with one job. See apps/web/DESIGN_SYSTEM.md.
 *
 * Components never name these directly — they use the semantic tokens
 * (--font-display, --font-numeric, --font-label, --font-body) that
 * tokens.css maps onto the variables below.
 *
 * Every face here is a pixel face, body copy included. That is a deliberate
 * choice for a pixel-art aesthetic, and it has a cost — see the type section
 * of DESIGN_SYSTEM.md before swapping anything.
 */

/** Titles, card headings, the lakeside sign, the logo. 18px and up. */
export const jersey15 = Jersey_15({
  weight: "400",
  subsets: ["latin"],
  display: "swap",
  variable: "--font-jersey-15",
});

/** Stat numbers, the timer readout, button labels. Same skeleton, more weight. */
export const jersey25 = Jersey_25({
  weight: "400",
  subsets: ["latin"],
  display: "swap",
  variable: "--font-jersey-25",
});

/** Uppercase micro-labels at 9-12px. Nothing longer than three words. */
export const silkscreen = Silkscreen({
  weight: ["400", "700"],
  subsets: ["latin"],
  display: "swap",
  variable: "--font-silkscreen",
});

/**
 * Body, emails, help text, inputs, errors.
 *
 * A Japanese bitmap gothic: even widths, real lowercase, and the one pixel
 * face that holds up across a full sentence. It needs more leading than a
 * normal sans — globals.css sets line-height 1.75 for exactly this reason.
 */
export const dotGothic = DotGothic16({
  weight: "400",
  subsets: ["latin"],
  display: "swap",
  variable: "--font-dotgothic",
});

export const fontVariables = [
  jersey15.variable,
  jersey25.variable,
  silkscreen.variable,
  dotGothic.variable,
].join(" ");

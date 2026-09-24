import { ElementType, HTMLAttributes } from "react";
import { cx } from "../../lib/cx";

export interface PixelPanelProps extends HTMLAttributes<HTMLElement> {
  /** Render as something other than a div — `section`, `article`, `form`. */
  as?: ElementType;
}

/**
 * The parchment surface. Every word in the app sits on one of these, never
 * directly on a scene background — that rule is what lets backgrounds be
 * swapped in later without a contrast pass on every screen.
 */
export function PixelPanel({
  as: Tag = "div",
  className,
  ...props
}: PixelPanelProps) {
  return <Tag className={cx("pixel-panel", className)} {...props} />;
}

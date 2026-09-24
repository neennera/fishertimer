import { HTMLAttributes } from "react";
import { cx } from "../../lib/cx";

/**
 * Currently one use: the ban flag on the account page (UC-07 ViewProfile).
 * A banned user still signs in — the ban is enforced at room join, not login.
 */
export function PixelBadge({
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement>) {
  return <span className={cx("pixel-badge", className)} {...props} />;
}

import { HTMLAttributes } from "react";
import { cx } from "../../lib/cx";

export function PixelBadge({
  className,
  ...props
}: HTMLAttributes<HTMLSpanElement>) {
  return <span className={cx("pixel-badge", className)} {...props} />;
}
